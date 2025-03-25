box.cfg{
  listen = 3301,
  log_level = 5,
  log = "tarantool.log"
}

-- Создаём space для хранения опросов
if not box.space.polls then
    box.schema.create_space('polls', {
        format = {
            {name = 'id', type = 'unsigned'},
            {name = 'question', type = 'string'},
            {name = 'options', type = 'array'},
            {name = 'votes', type = 'map'},
            {name = 'closed',   type = 'boolean'},
            {name = 'user_votes', type = 'map'},
        },
        if_not_exists = true
    })
    box.space.polls:create_index('primary', {
        parts = {'id'},
        if_not_exists = true
    })
end