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
