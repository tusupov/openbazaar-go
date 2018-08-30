package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/OpenBazaar/openbazaar-go/core"
	"github.com/OpenBazaar/openbazaar-go/pb"
	"github.com/OpenBazaar/openbazaar-go/repo"
	"github.com/OpenBazaar/openbazaar-go/test"
	"github.com/OpenBazaar/openbazaar-go/test/factory"
)

func TestSettings(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsJSON, checkIsSuccessJSONEqualTo(settingsJSON)},
		{"GET", "/ob/settings", "", checkIsSuccessJSONEqualTo(settingsJSON)},
		{"POST", "/ob/settings", settingsJSON, checkIsAlreadyExistsError},
		{"PUT", "/ob/settings", settingsUpdateJSON, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", checkIsSuccessJSONEqualTo(settingsUpdateJSON)},
		{"PUT", "/ob/settings", settingsUpdateJSON, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", checkIsSuccessJSONEqualTo(settingsUpdateJSON)},
		{"PATCH", "/ob/settings", settingsPatchJSON, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", checkIsSuccessJSONEqualTo(settingsPatchedJSON)},
	})
}

func TestSettingsInvalidPost(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsMalformedJSON, checkIs400Error},
	})
}

func TestSettingsInvalidPut(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsJSON, checkIsSuccessJSONEqualTo(settingsJSON)},
		{"GET", "/ob/settings", "", checkIsSuccessJSONEqualTo(settingsJSON)},
		{"PUT", "/ob/settings", settingsMalformedJSON, checkIs400Error},
	})
}

func TestProfile(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},
		{"POST", "/ob/profile", profileJSON, checkIsAlreadyExistsError},
		{"PUT", "/ob/profile", profileUpdateJSON, checkIsSuccessJSON},
		{"PUT", "/ob/profile", profileUpdatedJSON, checkIsSuccessJSON},
	})
}

func TestAvatar(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},
		{"POST", "/ob/avatar", avatarValidJSON, checkIsSuccessJSONEqualTo(avatarValidJSONResponse)},
	})
}

func TestAvatarNoProfile(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/avatar", avatarValidJSON, checkIs500Error(core.ErrorProfileNotFound)},
	})
}

func TestAvatarUnexpectedEOF(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},
		{"POST", "/ob/avatar", avatarUnexpectedEOFJSON, checkIs500Error(errors.New("unexpected EOF"))},
	})
}

func TestAvatarInvalidJSON(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},
		{"POST", "/ob/avatar", avatarInvalidTQJSON, checkIs500Error(errors.New("bad Tq value"))},
	})
}

func TestImages(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/images", imageValidJSON, checkIsSuccessJSONEqualTo(imageValidJSONResponse)},
	})
}

func TestHeader(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},
		{"POST", "/ob/header", headerValidJSON, checkIsSuccessJSONEqualTo(headerValidJSONResponse)},
	})
}

func TestHeaderNoProfile(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/header", headerValidJSON, checkIs500Error(core.ErrorProfileNotFound)},
	})
}

func TestModerator(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"PUT", "/ob/moderator", moderatorValidJSON, checkIs400Error},
	})

	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, checkIsSuccessJSON},

		// TODO: Enable after fixing bug that requires peers in order to set moderator status
		// {"PUT", "/ob/moderator", moderatorValidJSON, checkIsEmptyJSONObject},

		// // Update
		// {"PUT", "/ob/moderator", moderatorUpdatedValidJSON, checkIsEmptyJSONObject},
		{"DELETE", "/ob/moderator", "", checkIsEmptyJSONObject},
	})
}

func TestListings(t *testing.T) {
	t.Parallel()

	goodListingJSON := jsonFor(t, factory.NewListing("ron-swanson-tshirt"))
	updatedListing := factory.NewListing("ron-swanson-tshirt")
	updatedListing.Taxes = []*pb.Listing_Tax{
		{
			Percentage:  17,
			TaxShipping: true,
			TaxType:     "Sales tax",
			TaxRegions:  []pb.CountryCode{pb.CountryCode_UNITED_STATES},
		},
	}
	updatedListingJSON := jsonFor(t, updatedListing)

	runAPITests(t, []apiTest{
		{"GET", "/ob/listings", "", checkIsEmptyJSONArray},
		{"GET", "/ob/inventory", "", checkIsEmptyJSONObject},

		// Invalid creates
		{"POST", "/ob/listing", `{`, checkIs400Error},

		{"GET", "/ob/listings", "", checkIsEmptyJSONArray},
		{"GET", "/ob/inventory", "", checkIsEmptyJSONObject},

		// TODO: Add support for improved JSON matching to since contracts
		// change each test run due to signatures

		// Create/Get
		{"GET", "/ob/listing/ron-swanson-tshirt", "", checkIsNotFoundError},
		{"POST", "/ob/listing", goodListingJSON, checkIsSuccessJSONEqualTo(`{"slug": "ron-swanson-tshirt"}`)},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", checkIsSuccessJSON},
		{"POST", "/ob/listing", updatedListingJSON, checkIsAlreadyExistsError},

		// TODO: Add support for improved JSON matching to since contracts
		// change each test run due to signatures
		{"GET", "/ob/listings", "", checkIsSuccessJSON},

		// TODO: This returns `inventoryJSONResponse` but slices are unordered
		// so they don't get considered equal. Figure out a way to fix that.
		{"GET", "/ob/inventory", "", checkIsSuccessJSON},

		// Update inventory
		{"POST", "/ob/inventory", inventoryUpdateJSON, checkIsEmptyJSONObject},

		// Update/Get Listing
		{"PUT", "/ob/listing", updatedListingJSON, checkIsEmptyJSONObject},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", checkIsSuccessJSON},

		// Delete/Get
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", checkIsEmptyJSONObject},
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", checkIsNotFoundError},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", checkIsNotFoundError},

		// Mutate non-existing listings
		{"PUT", "/ob/listing", updatedListingJSON, checkIsNotFoundError},
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", checkIsNotFoundError},
	})
}

func TestCryptoListings(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	updatedListing := *listing

	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
		{"GET", "/ob/listing/crypto", jsonFor(t, &updatedListing), checkIsSuccessJSON},

		{"PUT", "/ob/listing", jsonFor(t, &updatedListing), checkIsEmptyJSONObject},
		{"PUT", "/ob/listing", jsonFor(t, &updatedListing), checkIsEmptyJSONObject},
		{"GET", "/ob/listing/crypto", jsonFor(t, &updatedListing), checkIsSuccessJSON},

		{"DELETE", "/ob/listing/crypto", "", checkIsEmptyJSONObject},
		{"DELETE", "/ob/listing/crypto", "", checkIsNotFoundError},
		{"GET", "/ob/listing/crypto", "", checkIsNotFoundError},
	})
}

func TestCryptoListingsPriceModifier(t *testing.T) {
	t.Parallel()

	outOfRangeErr := core.ErrPriceModifierOutOfRange{
		Min: core.PriceModifierMin,
		Max: core.PriceModifierMax,
	}

	listing := factory.NewCryptoListing("crypto")
	listing.Metadata.PriceModifier = core.PriceModifierMax
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
		{"GET", "/ob/listing/crypto", jsonFor(t, listing), checkIsSuccessJSON},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMax + 0.001
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMax + 0.01
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(outOfRangeErr)},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMin - 0.001
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMin - 1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(outOfRangeErr)},
	})
}

func TestListingsQuantity(t *testing.T) {
	t.Parallel()

	listing := factory.NewListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
	})

	listing.Item.Skus[0].Quantity = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSON},
	})

	listing.Item.Skus[0].Quantity = -1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSON},
	})
}

func TestCryptoListingsQuantity(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSONEqualTo(`{"slug": "crypto"}`)},
	})

	listing.Item.Skus[0].Quantity = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrCryptocurrencySkuQuantityInvalid)},
	})

	listing.Item.Skus[0].Quantity = -1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrCryptocurrencySkuQuantityInvalid)},
	})
}

func TestCryptoListingsNoCoinType(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	listing.Metadata.CoinType = ""

	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrCryptocurrencyListingCoinTypeRequired)},
	})
}

func TestCryptoListingsCoinDivisibilityIncorrect(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIsSuccessJSON},
	})

	listing.Metadata.CoinDivisibility = 1e7
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrListingCoinDivisibilityIncorrect)},
	})

	listing.Metadata.CoinDivisibility = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrListingCoinDivisibilityIncorrect)},
	})
}

func TestCryptoListingsIllegalFields(t *testing.T) {
	t.Parallel()

	runTest := func(listing *pb.Listing, err error) {
		runAPITests(t, []apiTest{
			{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(err)},
		})
	}

	physicalListing := factory.NewListing("physical")

	listing := factory.NewCryptoListing("crypto")
	listing.Metadata.PricingCurrency = "btc"
	runTest(listing, core.ErrCryptocurrencyListingIllegalField("metadata.pricingCurrency"))

	listing = factory.NewCryptoListing("crypto")
	listing.Item.Condition = "new"
	runTest(listing, core.ErrCryptocurrencyListingIllegalField("item.condition"))

	listing = factory.NewCryptoListing("crypto")
	listing.Item.Options = physicalListing.Item.Options
	runTest(listing, core.ErrCryptocurrencyListingIllegalField("item.options"))

	listing = factory.NewCryptoListing("crypto")
	listing.ShippingOptions = physicalListing.ShippingOptions
	runTest(listing, core.ErrCryptocurrencyListingIllegalField("shippingOptions"))

	listing = factory.NewCryptoListing("crypto")
	listing.Coupons = physicalListing.Coupons
	runTest(listing, core.ErrCryptocurrencyListingIllegalField("coupons"))
}

func TestMarketRatePrice(t *testing.T) {
	t.Parallel()

	listing := factory.NewListing("listing")
	listing.Metadata.Format = pb.Listing_Metadata_MARKET_PRICE
	listing.Item.Price = 1

	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), checkIs500Error(core.ErrMarketPriceListingIllegalField("item.price"))},
	})
}

func TestStatus(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/ob/status", "", checkIs400Error},
		{"GET", "/ob/status/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", "", checkIsSuccessJSON},
	})
}

func TestWallet(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/wallet/address", "", checkIsSuccessJSONEqualTo(walletAddressJSONResponse)},
		{"GET", "/wallet/balance", "", checkIsSuccessJSONEqualTo(walletBalanceJSONResponse)},
		{"GET", "/wallet/mnemonic", "", checkIsSuccessJSONEqualTo(walletMnemonicJSONResponse)},
		{"POST", "/wallet/spend", spendJSON, checkIs400Error},
	})
}

func TestConfig(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		// TODO: Need better JSON matching
		{"GET", "/ob/config", "", checkIsSuccessJSON},
	})
}

func Test404(t *testing.T) {
	t.Parallel()

	// Test undefined endpoints
	runAPITests(t, []apiTest{
		{"GET", "/ob/a", "{}", checkIsNotFoundError},
		{"PUT", "/ob/a", "{}", checkIsNotFoundError},
		{"POST", "/ob/a", "{}", checkIsNotFoundError},
		{"PATCH", "/ob/a", "{}", checkIsNotFoundError},
		{"DELETE", "/ob/a", "{}", checkIsNotFoundError},
	})
}

func TestPosts(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/ob/posts", "", checkIsEmptyJSONArray},

		// Invalid creates
		{"POST", "/ob/post", `{`, checkIs400Error},
		{"GET", "/ob/posts", "", checkIsEmptyJSONArray},

		// Create/Get
		{"GET", "/ob/post/test1", "", checkIsNotFoundError},
		{"POST", "/ob/post", postJSON, checkIsSuccessJSONEqualTo(postJSONResponse)},
		{"GET", "/ob/post/test1", "", checkIsSuccessJSON},
		{"POST", "/ob/post", postUpdateJSON, checkIsAlreadyExistsError},

		{"GET", "/ob/posts", "", checkIsSuccessJSON},

		// Update/Get Post
		{"PUT", "/ob/post", postUpdateJSON, checkIsEmptyJSONObject},
		{"GET", "/ob/post/test1", "", checkIsSuccessJSON},

		// Delete/Get
		{"DELETE", "/ob/post/test1", "", checkIsEmptyJSONObject},
		{"DELETE", "/ob/post/test1", "", checkIsNotFoundError},
		{"GET", "/ob/post/test1", "", checkIsNotFoundError},

		// Mutate non-existing listings
		{"PUT", "/ob/post", postUpdateJSON, checkIsNotFoundError},
		{"DELETE", "/ob/post/test1", "", checkIsNotFoundError},
	})
}

func TestCloseDisputeBlocksWhenExpired(t *testing.T) {
	t.Parallel()

	dbSetup := func(testRepo *test.Repository) error {
		expired := factory.NewExpiredDisputeCaseRecord()
		expired.CaseID = "expiredCase"
		for _, r := range []*repo.DisputeCaseRecord{expired} {
			if err := testRepo.DB.Cases().PutRecord(r); err != nil {
				return err
			}
			if err := testRepo.DB.Cases().UpdateBuyerInfo(r.CaseID, r.BuyerContract, []string{}, r.BuyerPayoutAddress, r.BuyerOutpoints); err != nil {
				return err
			}
		}
		return nil
	}
	expiredPostJSON := `{"orderId":"expiredCase","resolution":"","buyerPercentage":100.0,"vendorPercentage":0.0}`
	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"POST", "/ob/closedispute", expiredPostJSON, checkIs400Error},
	})
}

func TestZECSalesCannotReleaseEscrow(t *testing.T) {
	t.Parallel()

	sale := factory.NewSaleRecord()
	sale.Contract.VendorListings[0].Metadata.AcceptedCurrencies = []string{"ZEC"}
	dbSetup := func(testRepo *test.Repository) error {
		if err := testRepo.DB.Sales().Put(sale.OrderID, *sale.Contract, sale.OrderState, false); err != nil {
			return err
		}
		return nil
	}
	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"POST", "/ob/releaseescrow", fmt.Sprintf(`{"orderId":"%s"}`, sale.OrderID), checkIs400Error},
	})
}

func TestSalesGet(t *testing.T) {
	t.Parallel()

	sale := factory.NewSaleRecord()
	sale.Contract.VendorListings[0].Metadata.AcceptedCurrencies = []string{"BTC"}
	sale.Contract.VendorListings[0].Metadata.CoinType = "ZEC"
	sale.Contract.VendorListings[0].Metadata.ContractType = pb.Listing_Metadata_CRYPTOCURRENCY
	dbSetup := func(testRepo *test.Repository) error {
		return testRepo.DB.Sales().Put(sale.OrderID, *sale.Contract, sale.OrderState, false)
	}

	checkNewSale := func(t *testing.T, resp *httptest.ResponseRecorder) {
		if resp.Code != 200 {
			t.Fatal("Wanted status code 200 but got", resp.Code)
		}

		respObj := struct {
			Sales []repo.Sale `json:"sales"`
		}{}
		err := json.Unmarshal(resp.Body.Bytes(), &respObj)
		if err != nil {
			t.Fatal(err)
		}

		actualSale := respObj.Sales[0]

		if actualSale.BuyerHandle != sale.Contract.BuyerOrder.BuyerID.Handle {
			t.Fatal("Incorrect buyerHandle:", actualSale.BuyerHandle, "\nwanted:", sale.Contract.BuyerOrder.BuyerID.Handle)
		}
		if actualSale.BuyerId != sale.Contract.BuyerOrder.BuyerID.PeerID {
			t.Fatal("Incorrect buyerId:", actualSale.BuyerId, "\nwanted:", sale.Contract.BuyerOrder.BuyerID.PeerID)
		}
		if actualSale.CoinType != sale.Contract.VendorListings[0].Metadata.CoinType {
			t.Fatal("Incorrect coinType:", actualSale.CoinType, "\nwanted:", sale.Contract.VendorListings[0].Metadata.CoinType)
		}
		if actualSale.OrderId != sale.OrderID {
			t.Fatal("Incorrect orderId:", actualSale.OrderId, "\nwanted:", sale.OrderID)
		}
		if actualSale.PaymentCoin != sale.Contract.VendorListings[0].Metadata.AcceptedCurrencies[0] {
			t.Fatal("Incorrect paymentCoin:", actualSale.PaymentCoin, "\nwanted:", sale.Contract.VendorListings[0].Metadata.AcceptedCurrencies[0])
		}
		if actualSale.ShippingAddress != sale.Contract.BuyerOrder.Shipping.Address {
			t.Fatal("Incorrect shippingAddress:", actualSale.ShippingAddress, "\nwanted:", sale.Contract.BuyerOrder.Shipping.Address)
		}
		if actualSale.ShippingName != sale.Contract.BuyerOrder.Shipping.ShipTo {
			t.Fatal("Incorrect shippingName:", actualSale.ShippingName, "\nwanted:", sale.Contract.BuyerOrder.Shipping.ShipTo)
		}
		if actualSale.State != sale.OrderState.String() {
			t.Fatal("Incorrect state:", actualSale.State, "\nwanted:", sale.OrderState.String())
		}

	}

	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"GET", "/ob/sales", "", checkNewSale},
	})
}
func TestPurchasesGet(t *testing.T) {
	t.Parallel()

	purchase := factory.NewPurchaseRecord()
	purchase.Contract.VendorListings[0].Metadata.AcceptedCurrencies = []string{"BTC"}
	purchase.Contract.VendorListings[0].Metadata.CoinType = "ZEC"
	purchase.Contract.VendorListings[0].Metadata.ContractType = pb.Listing_Metadata_CRYPTOCURRENCY
	dbSetup := func(testRepo *test.Repository) error {
		return testRepo.DB.Purchases().Put(purchase.OrderID, *purchase.Contract, purchase.OrderState, false)
	}

	checkNewPurchase := func(t *testing.T, resp *httptest.ResponseRecorder) {
		if resp.Code != 200 {
			t.Fatal("Wanted status code 200 but got", resp.Code)
		}

		respObj := struct {
			Purchases []repo.Purchase `json:"purchases"`
		}{}
		err := json.Unmarshal(resp.Body.Bytes(), &respObj)
		if err != nil {
			t.Fatal(err)
		}

		actualPurchase := respObj.Purchases[0]

		if actualPurchase.VendorHandle != purchase.Contract.VendorListings[0].VendorID.Handle {
			t.Fatal("Incorrect vendorHandle:", actualPurchase.VendorId, "\nwanted:", purchase.Contract.VendorListings[0].VendorID.Handle)
		}
		if actualPurchase.VendorId != purchase.Contract.VendorListings[0].VendorID.PeerID {
			t.Fatal("Incorrect vendorId:", actualPurchase.VendorId, "\nwanted:", purchase.Contract.VendorListings[0].VendorID.PeerID)
		}
		if actualPurchase.CoinType != purchase.Contract.VendorListings[0].Metadata.CoinType {
			t.Fatal("Incorrect coinType:", actualPurchase.CoinType, "\nwanted:", purchase.Contract.VendorListings[0].Metadata.CoinType)
		}
		if actualPurchase.OrderId != purchase.OrderID {
			t.Fatal("Incorrect orderId:", actualPurchase.OrderId, "\nwanted:", purchase.OrderID)
		}
		if actualPurchase.PaymentCoin != purchase.Contract.VendorListings[0].Metadata.AcceptedCurrencies[0] {
			t.Fatal("Incorrect paymentCoin:", actualPurchase.PaymentCoin, "\nwanted:", purchase.Contract.VendorListings[0].Metadata.AcceptedCurrencies[0])
		}
		if actualPurchase.ShippingAddress != purchase.Contract.BuyerOrder.Shipping.Address {
			t.Fatal("Incorrect shippingAddress:", actualPurchase.ShippingAddress, "\nwanted:", purchase.Contract.BuyerOrder.Shipping.Address)
		}
		if actualPurchase.ShippingName != purchase.Contract.BuyerOrder.Shipping.ShipTo {
			t.Fatal("Incorrect shippingName:", actualPurchase.ShippingName, "\nwanted:", purchase.Contract.BuyerOrder.Shipping.ShipTo)
		}
		if actualPurchase.State != purchase.OrderState.String() {
			t.Fatal("Incorrect state:", actualPurchase.State, "\nwanted:", purchase.OrderState.String())
		}

	}

	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"GET", "/ob/purchases", "", checkNewPurchase},
	})
}

func TestCasesGet(t *testing.T) {
	t.Parallel()

	disputeCaseRecord := factory.NewDisputeCaseRecord()
	disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.AcceptedCurrencies = []string{"BTC"}
	disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.CoinType = "ZEC"
	disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.ContractType = pb.Listing_Metadata_CRYPTOCURRENCY
	disputeCaseRecord.CoinType = "ZEC"
	disputeCaseRecord.PaymentCoin = "BTC"
	dbSetup := func(testRepo *test.Repository) error {
		return testRepo.DB.Cases().PutRecord(disputeCaseRecord)
	}

	checkNewCase := func(t *testing.T, resp *httptest.ResponseRecorder) {
		if resp.Code != 200 {
			t.Fatal("Wanted status code 200 but got", resp.Code)
		}

		respObj := struct {
			Cases []repo.Case `json:"cases"`
		}{}
		err := json.Unmarshal(resp.Body.Bytes(), &respObj)
		if err != nil {
			t.Fatal(err)
		}

		actualCase := respObj.Cases[0]

		if actualCase.CoinType != disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.CoinType {
			t.Fatal("Incorrect coinType:", actualCase.CoinType, "\nwanted:", disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.CoinType)
		}
		if actualCase.PaymentCoin != disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.AcceptedCurrencies[0] {
			t.Fatal("Incorrect paymentCoin:", actualCase.PaymentCoin, "\nwanted:", disputeCaseRecord.BuyerContract.VendorListings[0].Metadata.AcceptedCurrencies[0])
		}
		if actualCase.State != disputeCaseRecord.OrderState.String() {
			t.Fatal("Incorrect state:", actualCase.State, "\nwanted:", disputeCaseRecord.OrderState.String())
		}

	}

	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"GET", "/ob/cases", "", checkNewCase},
	})
}

func TestNotificationsAreReturnedInExpectedOrder(t *testing.T) {
	t.Parallel()

	const sameTimestampsAreReturnedInReverse = `{
    "notifications": [
        {
            "notification": {
                "notificationId": "notif3",
                "peerId": "somepeerid",
                "type": "follow"
            },
            "read": false,
	    "timestamp": "1996-07-17T23:15:45Z",
            "type": "follow"
        },
        {
            "notification": {
                "notificationId": "notif2",
                "peerId": "somepeerid",
                "type": "follow"
            },
            "read": false,
	    "timestamp": "1996-07-17T23:15:45Z",
            "type": "follow"
        },
        {
            "notification": {
                "notificationId": "notif1",
                "peerId": "somepeerid",
                "type": "follow"
            },
            "read": false,
	    "timestamp": "1996-07-17T23:15:45Z",
            "type": "follow"
        }
    ],
    "total": 3,
    "unread": 3
}`

	const sameTimestampsAreReturnedInReverseAndRespectOffsetID = `{
    "notifications": [
        {
            "notification": {
                "notificationId": "notif2",
                "peerId": "somepeerid",
                "type": "follow"
            },
            "read": false,
            "timestamp": "1996-07-17T23:15:45Z",
            "type": "follow"
        },
        {
            "notification": {
                "notificationId": "notif1",
                "peerId": "somepeerid",
                "type": "follow"
            },
            "read": false,
            "timestamp": "1996-07-17T23:15:45Z",
            "type": "follow"
        }
    ],
    "total": 0,
    "unread": 3
}`
	var (
		createdAt = time.Unix(837645345, 0)
		notif1    = &repo.Notification{
			ID:           "notif1",
			CreatedAt:    createdAt,
			NotifierType: repo.NotifierTypeFollowNotification,
			NotifierData: &repo.FollowNotification{
				ID:     "notif1",
				Type:   repo.NotifierTypeFollowNotification,
				PeerId: "somepeerid",
			},
		}
		notif2 = &repo.Notification{
			ID:           "notif2",
			CreatedAt:    createdAt,
			NotifierType: repo.NotifierTypeFollowNotification,
			NotifierData: &repo.FollowNotification{
				ID:     "notif2",
				Type:   repo.NotifierTypeFollowNotification,
				PeerId: "somepeerid",
			},
		}
		notif3 = &repo.Notification{
			ID:           "notif3",
			CreatedAt:    createdAt,
			NotifierType: repo.NotifierTypeFollowNotification,
			NotifierData: &repo.FollowNotification{
				ID:     "notif3",
				Type:   repo.NotifierTypeFollowNotification,
				PeerId: "somepeerid",
			},
		}
	)
	dbSetup := func(testRepo *test.Repository) error {
		for _, n := range []*repo.Notification{notif1, notif2, notif3} {
			if err := testRepo.DB.Notifications().PutRecord(n); err != nil {
				return err
			}
		}
		return nil
	}

	runAPITestsWithSetup(t, dbSetup, []apiTest{
		{"GET", "/ob/notifications?limit=-1", "", checkIsSuccessJSONEqualTo(sameTimestampsAreReturnedInReverse)},
		{"GET", "/ob/notifications?limit=-1&offsetId=notif3", "", checkIsSuccessJSONEqualTo(sameTimestampsAreReturnedInReverseAndRespectOffsetID)},
	})
}

// TODO: Make NewDisputeCaseRecord return a valid fixture for this valid case to work
//func TestCloseDisputeReturnsOK(t *testing.T) {
//dbSetup := func(testRepo *test.Repository) error {
//nonexpired := factory.NewDisputeCaseRecord()
//nonexpired.CaseID = "nonexpiredCase"
//for _, r := range []*repo.DisputeCaseRecord{nonexpired} {
//if err := testRepo.DB.Cases().PutRecord(r); err != nil {
//return err
//}
//if err := testRepo.DB.Cases().UpdateBuyerInfo(r.CaseID, r.BuyerContract, []string{}, r.BuyerPayoutAddress, r.BuyerOutpoints); err != nil {
//return err
//}
//}
//return nil
//}
//nonexpiredPostJSON := `{"orderId":"nonexpiredCase","resolution":"","buyerPercentage":100.0,"vendorPercentage":0.0}`
//runAPITestsWithSetup(t, []apiTest{
//{"POST", "/ob/profile", moderatorProfileJSON, checkIsSuccessJSON},
//{"POST", "/ob/closedispute", nonexpiredPostJSON, checkIsSuccessJSON},
//}, dbSetup, nil)
//}
