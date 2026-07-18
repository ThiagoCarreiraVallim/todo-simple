-- Vincula listas a um usuário (opcional): listas anônimas continuam com user_id
-- NULL e acessíveis pelo slug (capability). Ao logar, as listas do usuário podem
-- ser recuperadas por user_id em qualquer aparelho. Aditivo e seguro para dados
-- já existentes em produção.
ALTER TABLE lists ADD COLUMN user_id uuid REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX lists_user_id_idx ON lists (user_id);
