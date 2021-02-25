package main

import (
	"context"
	"fmt"

	"github.com/onflow/cadence"
	flow "github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/client"
)

// format MomentPurchased(id: UInt64, price: UFix64, seller: Address?)
type MomentPurchasedEvent cadence.Event

// gives MomentPurhcasedEvent an Id() function, eg. evt.Id()
func (evt MomentPurchasedEvent) Id() uint64 {
	return uint64(evt.Fields[0].(cadence.UInt64))
}

func (evt MomentPurchasedEvent) Price() float64 {
	return float64(evt.Fields[1].(cadence.UFix64).ToGoValue().(uint64)) / 1e8
}

func (evt MomentPurchasedEvent) Seller() *flow.Address {
	optionalAddress := (evt.Fields[2].(cadence.Optional))
	if cadenceAddress, ok := optionalAddress.Value.(cadence.Address); ok {

		sellerAddress := flow.BytesToAddress(cadenceAddress.Bytes())
		return &sellerAddress
	}
	return nil
}

func GetSaleMomentFromOwnerAtBlock(flowClient *client.Client, blockHeight uint64, ownerAddress flow.Address, momentFlowID uint64) (*SaleMoment, error) {
	getSaleMomentScript := `
		import TopShot from 0x0b2a3299cc857e29
        import Market from 0xc1e4f4f4c4257510
        pub struct SaleMoment {
          pub var id: UInt64
          pub var playId: UInt32
          pub var play: {String: String}
          pub var setId: UInt32
          pub var setName: String
          pub var serialNumber: UInt32
          pub var price: UFix64
          pub var series: UInt32
          init(moment: &TopShot.NFT, price: UFix64) {
            self.id = moment.id
            self.playId = moment.data.playID
            self.play = TopShot.getPlayMetaData(playID: self.playId)!
            self.setId = moment.data.setID
            self.setName = TopShot.getSetName(setID: self.setId)!
            self.serialNumber = moment.data.serialNumber
            self.price = price
            self.series = TopShot.getSetSeries(setID: self.setId)!
          }
        }
		pub fun main(owner:Address, momentID:UInt64): SaleMoment {
			let acct = getAccount(owner)
            let collectionRef = acct.getCapability(/public/topshotSaleCollection)!.borrow<&{Market.SalePublic}>() ?? panic("Could not borrow capability from public collection")
			return SaleMoment(moment: collectionRef.borrowMoment(id: momentID)!,price: collectionRef.getPrice(tokenID: momentID)!)
		}
`
	res, err := flowClient.ExecuteScriptAtBlockHeight(context.Background(), blockHeight, []byte(getSaleMomentScript), []cadence.Value{
		cadence.BytesToAddress(ownerAddress.Bytes()),
		cadence.UInt64(momentFlowID),
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching sale moment from flow: %w", err)
	}
	saleMoment := SaleMoment(res.(cadence.Struct))
	return &saleMoment, nil
}

type SaleMoment cadence.Struct

func (s SaleMoment) ID() uint64 {
	return uint64(s.Fields[0].(cadence.UInt64))
}

func (s SaleMoment) PlayID() uint32 {
	return uint32(s.Fields[1].(cadence.UInt32))
}

func (s SaleMoment) SetName() string {
	return string(s.Fields[4].(cadence.String))
}

func (s SaleMoment) SetID() uint32 {
	return uint32(s.Fields[3].(cadence.UInt32))
}

func (s SaleMoment) Play() map[string]string {
	dict := s.Fields[2].(cadence.Dictionary)
	res := map[string]string{}
	for _, kv := range dict.Pairs {
		res[string(kv.Key.(cadence.String))] = string(kv.Value.(cadence.String))
	}
	return res
}

func (s SaleMoment) SerialNumber() uint32 {
	return uint32(s.Fields[5].(cadence.UInt32))
}

func (s SaleMoment) Price() float64 {
	return float64(s.Fields[6].(cadence.UFix64).ToGoValue().(uint64)) / 1e8
}

func (s SaleMoment) Series() uint32 {
	t := uint32(s.Fields[7].(cadence.UInt32))
	if t == 0 {
		return uint32(1)
	} else {
		return t
	}
}

func (s SaleMoment) String() string {
	playData := s.Play()
	// fmt.Println(playData)
	return fmt.Sprintf("Sale: %s %s #%d (%s Series %d) $%.2f -- Play ID: %d",
		playData["FullName"], playData["PlayCategory"], s.SerialNumber(), s.SetName(), s.Series(), s.Price(), s.PlayID())
}
