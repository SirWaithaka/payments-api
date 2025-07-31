package payments

const (
	BankMpesa       = "W001"
	BankAirtelMoney = "W002"
)

type Bank struct {
	Name string
	Code string
}

var Banks = []Bank{
	{Name: "MPESA", Code: "W001"},
	{Name: "AIRTEL MONEY", Code: "W002"},
}
