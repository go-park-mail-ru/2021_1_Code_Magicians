package entity

type key string

const CookieInfoKey key = "cookieInfo"
const CookieNameKey key = "session_id"

const VkAuthURLKey key = "https://oauth.vk.com/"
const VkAPIURLKey key = "https://api.vk.com/"
const VkAuthenticateURLKey key = "https://pinterbest.ru/login/callback"
const VkAddTokenURLKey key = "https://pinterbest.ru/add_vk/callback"
const VkCreateUserURLKey key = "https://pinterbest.ru/signup/callback"

const IDKey key = "id"
const UsernameKey key = "username"
const SearchKeyQuery key = "searchKey"

const UserAvatarDefaultPath key = "assets/img/default-avatar.jpg"
const BoardAvatarDefaultPath key = "assets/img/default-board-avatar.jpg"

const AllNotificationsTypeKey key = "all-notifications"
const OneNotificationTypeKey key = "notification"

const AllChatsTypeKey key = "all-chats"
const OneChatTypeKey key = "new-chat"
const OneMessageTypeKey key = "new-message"

const PinInfoLabelKey key = "pinInfo"
const PinImageLabelKey key = "pinImage"
const PinIDLabelKey key = "pinID"
const PinAmountLabelKey key = "num"

const EmailTemplateFilenameKey key = "pin_email_template.html"
