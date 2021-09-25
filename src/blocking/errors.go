package blocking

import "errors"


var (
    ErrCidrAlreadyExists = errors.New( "IP range already blocked; cannot be overwritten" )
)
