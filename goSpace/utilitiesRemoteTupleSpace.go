package tuplespace

import (
	"encoding/gob"
	"fmt"
	"net"
	"reflect"
	"svn/bachelorProject/tupleSpaceFramework/constants"
	"svn/bachelorProject/tupleSpaceFramework/topology"
)

// Put will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and tuple specified by the user.
// The method returns a boolean to inform if the operation was carried out with
// success or not.
func Put(ptp topology.PointToPoint, tupleFields ...interface{}) bool {
	t := CreateTuple(tupleFields)
	conn, errDial := establishConnection(ptp)

	// Error check for establishing connection.
	if errDial != nil {
		fmt.Println("ErrDial:", errDial)
		return false
	}

	// Make sure the connection closes when method returns.
	defer conn.Close()

	errSendMessage := sendMessage(conn, constants.PutRequest, t)

	// Error check for sending message.
	if errSendMessage != nil {
		fmt.Println("ErrSendMessage:", errSendMessage)
		return false
	}

	b, errReceiveMessage := receiveMessageBool(conn)

	// Error check for receiving response.
	if errReceiveMessage != nil {
		fmt.Println("ErrReceiveMessage:", errReceiveMessage)
		return false
	}

	// Return result.
	return b
}

// PutP will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and tuple specified by the user.
// As the method is nonblocking it wont wait for a response whether or not the
// operation was successful.
func PutP(t Tuple, ptp topology.PointToPoint) {
	conn, errDial := establishConnection(ptp)

	// Error check for establishing connection.
	if errDial != nil {
		fmt.Println("ErrDial:", errDial)
	}

	// Make sure the connection closes when method returns.
	defer conn.Close()

	errSendMessage := sendMessage(conn, constants.PutPRequest, t)

	// Error check for sending message.
	if errSendMessage != nil {
		fmt.Println("ErrSendMessage:", errSendMessage)
	}
}

// Get will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and template specified by the user.
// The method returns a tuple that matches the template.
func Get(ptp topology.PointToPoint, tempFields ...interface{}) {
	getAndQuery(tempFields, ptp, constants.GetRequest)
}

// Query will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and template specified by the user.
// The method returns a tuple that matches the template.
func Query(ptp topology.PointToPoint, tempFields ...interface{}) {
	getAndQuery(tempFields, ptp, constants.QueryRequest)
}

func getAndQuery(tempFields []interface{}, ptp topology.PointToPoint, operation string) {
	t := CreateTemplate(tempFields)
	conn, errDial := establishConnection(ptp)

	// Error check for establishing connection.
	if errDial != nil {
		fmt.Println("ErrDial:", errDial)
	}

	// Make sure the connection closes when method returns.
	defer conn.Close()

	errSendMessage := sendMessage(conn, operation, t)

	// Error check for sending message.
	if errSendMessage != nil {
		fmt.Println("ErrSendMessage:", errSendMessage)
	}

	tuple, errReceiveMessage := receiveMessageTuple(conn)

	// Error check for receiving response.
	if errReceiveMessage != nil {
		fmt.Println("ErrReceiveMessage:", errReceiveMessage)
	}
	WriteTupleToVariables(tuple, tempFields)
	// Return result.
	return
}

// GetP will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and template specified by the user.
// The method is nonblocking and will return a boolean that specifies if a
// tuple matching the template was found or not and the tuple if it found any.
func GetP(ptp topology.PointToPoint, tempFields ...interface{}) (bool, Tuple) {
	return getPAndQueryP(tempFields, ptp, constants.GetPRequest)
}

// QueryP will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation and template specified by the user.
// The method is nonblocking and will return a boolean that specifies if a
// tuple matching the template was found or not and the tuple if it found any.
func QueryP(ptp topology.PointToPoint, tempFields ...interface{}) (bool, Tuple) {
	return getPAndQueryP(tempFields, ptp, constants.QueryPRequest)
}

func getPAndQueryP(tempFields []interface{}, ptp topology.PointToPoint, operation string) (bool, Tuple) {
	t := CreateTemplate(tempFields)
	conn, errDial := establishConnection(ptp)

	// Error check for establishing connection.
	if errDial != nil {
		fmt.Println("ErrDial:", errDial)
	}

	// Make sure the connection closes when method returns.
	defer conn.Close()

	errSendMessage := sendMessage(conn, operation, t)

	// Error check for sending message.
	if errSendMessage != nil {
		fmt.Println("ErrSendMessage:", errSendMessage)
	}

	b, tuple, errReceiveMessage := receiveMessageBoolAndTuple(conn)

	// Error check for receiving response.
	if errReceiveMessage != nil {
		fmt.Println("ErrReceiveMessage:", errReceiveMessage)
	}
	if b {
		WriteTupleToVariables(tuple, tempFields)
	}

	// Return result.
	return b, tuple
}

// GetAll will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation specified by the user.
// The method is nonblocking and will return all tuples found in the tuple
// space.
// NOTE: tuples is allowed to be an empty list, implying the tuple space was
// empty.
func GetAll(ptp topology.PointToPoint) []Tuple {
	return getAllAndQueryAll(ptp, constants.GetAllRequest)
}

// QueryAll will open a TCP connection to the PointToPoint and send the message,
// which includes the type of operation specified by the user.
// The method is nonblocking and will return all tuples found in the tuple
// space.
// NOTE: tuples is allowed to be an empty list, implying the tuple space was
// empty.
func QueryAll(ptp topology.PointToPoint) []Tuple {
	return getAllAndQueryAll(ptp, constants.QueryAllRequest)
}

func getAllAndQueryAll(ptp topology.PointToPoint, operation string) []Tuple {
	conn, errDial := establishConnection(ptp)

	// Error check for establishing connection.
	if errDial != nil {
		fmt.Println("ErrDial:", errDial)
	}

	// Make sure the connection closes when method returns.
	defer conn.Close()

	// Initiallise dummy tuple.
	// TODO: Get rid of the dummy tuple.
	t := Tuple{}
	errSendMessage := sendMessage(conn, operation, t)

	// Error check for sending message.
	if errSendMessage != nil {
		fmt.Println("ErrSendMessage:", errSendMessage)
	}

	tuples, errReceiveMessage := receiveMessageTupleList(conn)

	// Error check for receiving response.
	if errReceiveMessage != nil {
		fmt.Println("ErrReceiveMessage:", errReceiveMessage)
	}

	// Return result.
	return tuples
}

// establishConnection will establish a connection to the PointToPoint ptp and
// return the Conn and error.
func establishConnection(ptp topology.PointToPoint) (net.Conn, error) {
	addr := ptp.GetAddress()

	// Establish a connection to the PointToPoint using TCP to ensure reliability.
	conn, errDial := net.Dial("tcp", addr)

	return conn, errDial
}

func sendMessage(conn net.Conn, operation string, t interface{}) error {
	// Create encoder to the connection.
	enc := gob.NewEncoder(conn)

	// Register the type of t for Encode to handle it.
	gob.Register(t)
	//registrer typefield to match types
	gob.Register(TypeField{})

	// Generate the message.
	message := topology.CreateMessage(operation, t)

	// Sends the message to the connection through the encoder.
	errEnc := enc.Encode(message)

	return errEnc
}

func receiveMessageBool(conn net.Conn) (bool, error) {
	// Create decoder to the connection to receive the response.
	dec := gob.NewDecoder(conn)

	// Read the response from the connection through the decoder.
	var b bool
	errDec := dec.Decode(&b)

	return b, errDec
}

func receiveMessageTuple(conn net.Conn) (Tuple, error) {
	// Create decoder to the connection to receive the response.
	dec := gob.NewDecoder(conn)

	// Read the response from the connection through the decoder.
	var tuple Tuple
	errDec := dec.Decode(&tuple)

	return tuple, errDec
}

func receiveMessageBoolAndTuple(conn net.Conn) (bool, Tuple, error) {
	// Create decoder to the connection to receive the response.
	dec := gob.NewDecoder(conn)

	// Read the response from the connection through the decoder.
	var result []interface{}
	errDec := dec.Decode(&result)

	// Extract the boolean and tuple from the result.
	b := result[0].(bool)
	tuple := result[1].(Tuple)

	return b, tuple, errDec
}

func receiveMessageTupleList(conn net.Conn) ([]Tuple, error) {
	// Create decoder to the connection to receive the response.
	dec := gob.NewDecoder(conn)

	// Read the response from the connection through the decoder.
	var tuples []Tuple
	errDec := dec.Decode(&tuples)

	return tuples, errDec
}

// WriteTupleToVariables will overwrite the value of pointers in varibles, to
// the value in the tuple
func WriteTupleToVariables(t Tuple, variables []interface{}) {
	for i, value := range variables {
		if reflect.TypeOf(value).Kind() == reflect.Ptr {
			//Yep, this is how you change the value of a pointer
			reflect.ValueOf(value).Elem().Set(reflect.ValueOf(t.GetFieldAt(i)))
		}
	}
}