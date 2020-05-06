package ftp

// FTP Reply codes (see RFC 959 section 4.2)
const (
	ReplyFileStatusOK       = "150 File status okay; about to open data connection."
	ReplyCmdOK              = "200 Command okay."
	ReplySystemStatus       = "211 %s"
	ReplyFileStatus         = "213 %s"
	ReplySystemType         = "215 UNIX System type."
	ReplyServiceReady       = "220 Service ready for new user."
	ReplyClosingConn        = "221 Service closing control connection."
	ReplyClosingDataConn    = "226 Closing data connection."
	ReplyEnteringPasv       = "227 Entering Passive Mode (%d,%d,%d,%d,%d,%d)"
	ReplyUserLoggedIn       = "230 User logged in proceed."
	ReplyFileActionOK       = "250 Requested file action okay completed."
	ReplyPathNameOK         = "257 %s"
	ReplyUserNameOK         = "331 User name okay need password."
	ReplyCannotOpenDataConn = "425 Can't open data connection."
	ReplyActionAborted      = "450 Requested file action aborted."
	ReplyNotImplemented     = "502 Command not implemented."
	ReplyFileUnavailable    = "550 File unavailable."
)
