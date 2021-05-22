box.cfg{listen = 3301}
box.schema.user.passwd('pass')

function restore_sessions()
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

pcall(restore_sessions)