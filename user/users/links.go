package users

import (
	"flag"
	"fmt"
	"os"
)

var (
	domain    string
	entitymap = map[string]string{
		"customer": "customers",
		"address":  "addresses",
		"card":     "cards",
	}
)

func init() {
	flag.StringVar(&domain, "link-domain", os.Getenv("HATEAOS"), "HATEAOS link domain")
}

type Href struct {
	string `json:"href"`
}

type Links map[string]Href

func (l *Links) AddLink(ent string, id string) {
	link := fmt.Sprintf("http://%v/%v/%v", domain, entitymap[ent], id)
	(*l)[ent] = Href{link}
	(*l)["self"] = Href{link}
}

func (l *Links) AddAttrLink(attr string, corent string, id string) {
	link := fmt.Sprintf("http://%v/%v/%v/%v", domain, entitymap[corent], id, entitymap[attr])
	(*l)[entitymap[attr]] = Href{link}
}

func (l *Links) AddCustomer(id string) {
	l.AddLink("customer", id)
	l.AddAttrLink("address", "customer", id)
	l.AddAttrLink("card", "customer", id)
}

func (l *Links) AddAddress(id string) {
	l.AddLink("address", id)
}

func (l *Links) AddCard(id string) {
	l.AddLink("card", id)
}
