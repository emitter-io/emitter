package security

// Contract represents an interface for a contract.
type Contract interface {
	Validate(key Key) bool // Validate checks the security key with the contract.
}

// ContractData gathers data about the contract.
type ContractData struct {
	ID            int32 // Gets or sets the contract id.
	MasterID      uint16
	EncryptionKey string // Gets or sets the encryption key.
	Signature     int32  // Gets or sets the signature of the contract.
}

// Validate validates the contract data against a key.
func (c *ContractData) Validate(key Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract()
}

// ContractProvider represents an interface for a contract provider.
type ContractProvider interface {
	// Creates a new instance of a Contract in the underlying data storage.
	Create() *ContractData
	GetById(id int32) *ContractData
}

//SingleContractProvider provide contracts on premise.
type SingleContractProvider struct {
	Data *ContractData
}

// Create a contract, the SingleContractProvider way.
func (p *SingleContractProvider) Create(license *License) *ContractData {
	p.Data = new(ContractData)
	p.Data.MasterID = 1
	p.Data.ID = license.Contract
	p.Data.Signature = license.Signature

	return p.Data
}

// GetByID returns a ContractData fetched by its id.
func (p *SingleContractProvider) GetByID(id int32) *ContractData {
	if p.Data == nil || p.Data.ID != id {
		return nil
	}
	return p.Data
}
