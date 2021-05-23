box.cfg{listen = 3301}
box.schema.user.passwd('pass')

function restore_sessions_schema()
    s = box.schema.space.create('sessions')
    s:format({
             {name = 'user_id', type = 'unsigned'},
             {name = 'session_value', type = 'string'},
             {name = 'expiration_date', type = 'unsigned'}
             })
    s:create_index('primary', {
             type = 'hash',
             parts = {'user_id'}
             })
    s:create_index('secondary', {
             type = 'hash',
             parts = {'session_value'},
             unique= true
             })
end

pcall(restore_sessions_schema)

function restore_notifications_schema()
    s = box.schema.space.create('notifications')
    s:format({
             {name = 'notification_id', type = 'unsigned'},
             {name = 'user_id', type = 'unsigned'},
             {name = 'category', type = 'string'},
             {name = 'title', type = 'string'},
             {name = 'text', type = 'string'},
             {name = 'is_read', type = 'boolean'},
             })
    s:create_index('primary', {
             type = 'hash',
             parts = {'notification_id'}
             })
    s:create_index('secondary', {
             type = 'hash',
             parts = {'user_id'},
             })
end

pcall(restore_notifications_schema)
