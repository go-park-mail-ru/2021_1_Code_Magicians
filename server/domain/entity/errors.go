package entity

type customError string

const TransactionBeginError customError = "Could not start transaction"
const TransactionCommitError customError = "Could not commit transaction"

const DuplicatingCookieValueError customError = "Cookie with such session value already exists"
const UserNotLoggedInError customError = "User is not logged in"

const UserNotFoundError customError = "User not found"
const UsersNotFoundError customError = "Users not found"
const UsernameEmailDuplicateError customError = "Username or email is already taken"
const IncorrectPasswordError customError = "Password is incorrect"
const UserSavingError customError = "User saving failed"
const ValidationError customError = "Validation error"
const UnauthorizedError customError = "Unauthorized"

const FollowNotFoundError customError = "Follow relation not found"
const FollowAlreadyExistsError customError = "Follow relation already exists"
const FollowCountUpdateError customError = "Failed to update follow(er/ing) counter"
const SelfFollowError customError = "Users can't follow themselves"

const CookieGenerationError customError = "Could not generate cookie"
const GetCookieFromContextError customError = "Could not get cookie from context"

const FilenameGenerationError customError = "Could not generate filename"
const FileUploadError customError = "File upload failed"
const FileDeletionError customError = "File deletion failed"

const ClientNotSetError customError = "Websocket client not set"

const NoNotificationsError customError = "No notifications found"
const NotificationNotFoundError customError = "Notification not found"
const NotificationAlreadyReadError customError = "Notification was already read"

const ChatNotFoundError customError = "Chat not found"
const MessageNotFoundError customError = "Message not found"
const ChatAlreadyExistsError customError = "Chat already exists"
const UserNotInChatError customError = "User is not in chat"
const ChatAlreadyReadError customError = "Chat is already read"

const JsonMarshallError customError = "Could not parse struct into JSON"

const NotFoundInitUserBoard customError = "Could not find user's initial board"
const DeleteBoardError customError = "Could not delete user's board"
const CreateBoardError customError = "Could not create board"
const BoardNotFoundError customError = "No board found"
const GetBoardsByUserIDError customError = "No boards found in database with passed userID"
const DeleteInitBoardError customError = "Can not delete user's first board"
const CheckBoardOwnerError customError = "That board is not associated with that user"

const DeletePinError customError = "Could not delete pin"
const RemovePinError customError = "Could not remove pin from board"
const CreatePinError customError = "Could not create pin"
const AddPinToBoardError customError = "Could not add pin to board"
const GetPinReferencesCount customError = "Could not count the number of pin references"
const PinNotFoundError customError = "No pin found"
const GetPinsByBoardIdError customError = "Could not get pins from passed board"
const PinSavingError customError = "Pin saving failed"
const FeedLoadingError customError = "Could not extract pins for feed"
const NonPositiveNumOfPinsError customError = "Cannot get negative amount of pins"

const SearchingError customError = "Could not get results of searching"
const NoResultSearch customError = "Not results for search"

const AddCommentError customError = "Comment creation failed"
const GetCommentsError customError = "Could not get comments"
const ReturnCommentsError customError = "Could not return comments"

const NoPicturePassed customError = "No picture was passed"
const TooLargePicture customError = "Picture is too large"

func (err customError) Error() string { // customError implements error interface
	return string(err)
}
