package users

// User é a identidade do app: login sem senha por nome de usuário. O id (uuid)
// é a chave usada para vincular fazenda e listas.
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
