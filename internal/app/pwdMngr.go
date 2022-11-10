package app

import "golang.org/x/crypto/bcrypt"

type pwdMngr struct {
	cost int
}

func newPwdMngr(cost int) *pwdMngr {
	return &pwdMngr{cost}
}

func (p *pwdMngr) hash(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, p.cost)
}

func (p *pwdMngr) compare(hash, password []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}
