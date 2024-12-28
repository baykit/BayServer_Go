package common

import "bayserver-core/baykit/bayserver/util/exception"

/****************************************/
/* Interface Recipient                  */
/****************************************/

// Recipient
//   Letter receiver

type Recipient interface {

	// Receive
	//    Receives letters
	Receive(wait bool) (bool, exception.IOException)

	// Wakeup
	//   Wakes up the recipient
	Wakeup()

	// End
	//   Ends the recipient
	End()
}
