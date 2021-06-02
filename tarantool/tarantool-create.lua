box.cfg{listen = 3301}
box.schema.user.passwd('pass')

function restore_sessions_schema()
    sessions = box.schema.space.create('sessions')
    sessions:format({
             {name = 'user_id', type = 'unsigned'},
             {name = 'session_value', type = 'string'},
             {name = 'expiration_date', type = 'unsigned'}
             })
    sessions:create_index('primary', {
             type = 'tree',
             parts = {'user_id'},
             unique = true
             })
    sessions:create_index('secondary', {
             type = 'tree',
             parts = {'session_value'},
             unique= true
             })
end

pcall(restore_sessions_schema)

function restore_vk_tokens_schema()
    sessions = box.schema.space.create('vk_tokens')
    sessions:format({
             {name = 'user_id', type = 'unsigned'},
             {name = 'vk_token', type = 'string'},
             {name = 'expiration_date', type = 'unsigned'}
             })
    sessions:create_index('primary', {
             type = 'tree',
             parts = {'user_id'},
             unique = true
             })
    sessions:create_index('secondary', {
             type = 'tree',
             parts = {'vk_token'},
             unique = true
             })
end

pcall(restore_vk_tokens_schema)

function restore_notifications_schema()
    notifications = box.schema.space.create('notifications')
    notifications:format({
             {name = 'notification_id', type = 'unsigned'},
             {name = 'user_id', type = 'unsigned'},
             {name = 'category', type = 'string'},
             {name = 'title', type = 'string'},
             {name = 'text', type = 'string'},
             {name = 'is_read', type = 'boolean'},
             })

    box.schema.sequence.create('notification_id_sequence')
    notifications:create_index('primary', {
             type = 'tree',
             parts = {'notification_id'},
             sequence = 'notification_id_sequence',
             unique = true
             })
    notifications:create_index('secondary', {
             type = 'tree',
             parts = {'user_id'},
             unique = false
             })
end

pcall(restore_notifications_schema)

function restore_chats_schema()
    chats = box.schema.space.create('chats')
    chats:format({
             {name = 'chat_id', type = 'unsigned'},
             {name = 'first_user_id', type = 'unsigned'},
             {name = 'second_user_id', type = 'unsigned'},
             {name = 'first_user_read', type = 'boolean'},
             {name = 'second_user_read', type = 'boolean'},
             })

    box.schema.sequence.create('chat_id_sequence')
    chats:create_index('primary', {
             type = 'tree',
             parts = {'chat_id'},
             sequence = 'chat_id_sequence',
             unique = true
             })

    chats:create_index('secondary', {
             type = 'tree',
             parts = {'first_user_id', 'second_user_id'},
             unique = true
             })

    chats:create_index('by_first_user', {
             type = 'tree',
             parts = {'first_user_id'},
             unique = false
             })
    chats:create_index('by_second_user', {
             type = 'tree',
             parts = {'second_user_id'},
             unique = false
             })
end

pcall(restore_chats_schema)

function restore_messages_schema()
    messages = box.schema.space.create('messages')
    messages:format({
             {name = 'message_id', type = 'unsigned'},
             {name = 'chat_id', type = 'unsigned'},
             {name = 'author_id', type = 'unsigned'},
             {name = 'text', type = 'string'},
             {name = 'creation_time', type = 'string'},
             })

    box.schema.sequence.create('message_id_sequence')
    messages:create_index('primary', {
             type = 'tree',
             parts = {'message_id'},
             sequence = 'message_id_sequence',
             unique = true
             })
    messages:create_index('secondary', {
             type = 'tree',
             parts = {'chat_id'},
             unique = false
             })
end

pcall(restore_messages_schema)