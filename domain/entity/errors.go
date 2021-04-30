package entity

type customError string

const TransactionBeginError customError = "Could not start transaction"
const TransactionCommitError customError = "Could not commit transaction"

const UserNotFoundError customError = "User not found"
const UsernameEmailDuplicateError customError = "Username or email is already taken"
const IncorrectPasswordError customError = "Password is incorrect"
const UserSavingError customError = "User saving failed"
const ValidationError customError = "Validation error"
const UnauthorizedError customError = "Unauthorized"

const FollowNotFoundError customError = "Follow relation not found"
const FollowAlreadyExistsError customError = "Follow relation already exists"
const FollowCountUpdateError customError = "Failed to update follow(er/ing) counter"

const CookieGenerationError customError = "Could not generate cookie"
const GetCookieFromContextError customError = "Could not get cookie from context"

const FilenameGenerationError customError = "Could not generate filename"
const FileUploadError customError = "File upload failed"
const FileDeletionError customError = "File deletion failed"

const NoNotificationsError customError = "No notifications found"
const NotificationNotFoundError customError = "Notification not found"
const NotificationsClientNotSetError customError = "Notifications client not set"
const NotificationAlreadyReadError customError = "Notification was already read"
const NotificationSendError customError = "Notification send error"

const CreateBoardError customError = "Could not create board"
const GetBoardError customError = "No board found"
const GetBoardsByUserIDError customError = "User's boards not found"
const DeleteInitBoardError customError = "Can not delete user's first Board"
const NoPicturePassed customError = "No picture was passed"
const TooLargePicture customError = "Picture is too large"
//const IncorrectPasswordError customError = "Password is incorrect"
//const UserSavingError customError = "User saving failed"
//
//const FollowNotFoundError customError = "Follow relation not found"
//const FollowAlreadyExistsError customError = "Follow relation already exists"
//const FollowCountUpdateError customError = "Failed to update follow(er/ing) counter"
//
//const CookieGenerationError customError = "Could not generate cookie"
//
//const FilenameGenerationError customError = "Could not generate filename"
//const FileUploadError customError = "File upload failed"
//const FileDeletionError customError = "File deletion failed"
//
//const NoNotificationsError customError = "No notifications found"
//const NotificationNotFoundError customError = "Notification not found"
//const NotificationsClientNotSetError customError = "Notifications client not set"
//const NotificationAlreadyReadError customError = "Notification was already read"

// TODO: errors for pins, boards and comments

func (err customError) Error() string { // customError implements error interface
	return string(err)
}
