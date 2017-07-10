package security

// Contract represents an interface for a contract.
type Contract interface {
	Validate(key Key) bool // Validate checks the security key with the contract.
}

// contract represents a contract (user account).
type contract struct {
	ID            int32  // Gets or sets the contract id.
	MasterID      uint16 // Gets or sets the master id.
	EncryptionKey string // Gets or sets the encryption key.
	Signature     int32  // Gets or sets the signature of the contract.
}

// Validate validates the contract data against a key.
func (c *contract) Validate(key Key) bool {
	return c.MasterID == key.Master() &&
		c.Signature == key.Signature() &&
		c.ID == key.Contract()
}

// ContractProvider represents an interface for a contract provider.
type ContractProvider interface {
	// Creates a new instance of a Contract in the underlying data storage.
	Create() Contract
	Get(id int32) Contract
}

//SingleContractProvider provide contracts on premise.
type SingleContractProvider struct {
	Data *contract
}

// Create a contract, the SingleContractProvider way.
func (p *SingleContractProvider) Create(license *License) Contract {
	p.Data = new(contract)
	p.Data.MasterID = 1
	p.Data.ID = license.Contract
	p.Data.Signature = license.Signature

	return p.Data
}

// Get returns a ContractData fetched by its id.
func (p *SingleContractProvider) Get(id int32) Contract {
	if p.Data == nil || p.Data.ID != id {
		return nil
	}
	return p.Data
}
