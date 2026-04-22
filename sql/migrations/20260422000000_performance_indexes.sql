-- +goose Up
-- Busca inversa: encontrar usuários que têm um jogo específico
CREATE INDEX idx_user_games_game_id ON user_games(game_id, user_id);

-- Lookup de links pelo lado do recipient (NOT EXISTS no match)
CREATE INDEX idx_user_links_recipient ON user_links(recipient_id, requester_id);

-- +goose Down
DROP INDEX idx_user_games_game_id;
DROP INDEX idx_user_links_recipient;
