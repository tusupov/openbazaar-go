package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/OpenBazaar/openbazaar-go/core"
	"github.com/OpenBazaar/openbazaar-go/pb"
	"github.com/OpenBazaar/openbazaar-go/repo"
	"github.com/OpenBazaar/openbazaar-go/test"
	"github.com/OpenBazaar/openbazaar-go/test/factory"
)

func checkIsEqualJSON(expected string) checkFn {
	return func(t *testing.T, respBody []byte) {
		var responseJSON interface{}
		err := json.Unmarshal(respBody, &responseJSON)
		if err != nil {
			t.Fatal(err)
		}

		var expectedJSON interface{}
		err = json.Unmarshal([]byte(expected), &expectedJSON)
		if err != nil {
			fmt.Println(expected)
			t.Fatal(err)
		}

		if !reflect.DeepEqual(responseJSON, expectedJSON) {
			t.Error("Incorrect response.\nWanted:", expected, "\nGot:", string(respBody))
		}
	}
}

func checkIsJSON(t *testing.T, respBody []byte) {
	respBodyStr := string(respBody)
	if !(strings.HasPrefix(respBodyStr, "{") && strings.HasSuffix(respBodyStr, "}")) {
		t.Fatal("Response is not JSON:", respBodyStr)
	}
}

func checkIsEmptyJSONObject(t *testing.T, respBody []byte) {
	respBodyStr := string(respBody)
	if respBodyStr != "{}" {
		t.Fatal("Response is not empty JSON object:", respBodyStr)
	}
}

func checkIsEmptyJSONArray(t *testing.T, respBody []byte) {
	respBodyStr := string(respBody)
	if respBodyStr != "[]" {
		t.Fatal("Response is not empty JSON array:", respBodyStr)
	}
}

func checkIsErrorResponseJSON(errStr string) checkFn {
	return checkIsEqualJSON(`{"success": false, "reason": "` + errStr + `"}`)
}

func TestSettings(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsJSON, 200, checkIsEqualJSON(settingsJSON)},
		{"GET", "/ob/settings", "", 200, checkIsEqualJSON(settingsJSON)},
		{"POST", "/ob/settings", settingsJSON, 409, checkIsEqualJSON(settingsAlreadyExistsJSON)},
		{"PUT", "/ob/settings", settingsUpdateJSON, 200, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", 200, checkIsEqualJSON(settingsUpdateJSON)},
		{"PUT", "/ob/settings", settingsUpdateJSON, 200, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", 200, checkIsEqualJSON(settingsUpdateJSON)},
		{"PATCH", "/ob/settings", settingsPatchJSON, 200, checkIsEmptyJSONObject},
		{"GET", "/ob/settings", "", 200, checkIsEqualJSON(settingsPatchedJSON)},
	})
}

func TestSettingsInvalidPost(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsMalformedJSON, 400, checkIsEqualJSON(settingsMalformedJSONResponse)},
	})
}

func TestSettingsInvalidPut(t *testing.T) {
	t.Parallel()
	runAPITests(t, []apiTest{
		{"POST", "/ob/settings", settingsJSON, 200, checkIsEqualJSON(settingsJSON)},
		{"GET", "/ob/settings", "", 200, checkIsEqualJSON(settingsJSON)},
		{"PUT", "/ob/settings", settingsMalformedJSON, 400, checkIsEqualJSON(settingsMalformedJSONResponse)},
	})
}

func TestProfile(t *testing.T) {
	t.Parallel()

	// Create, Update
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},
		{"POST", "/ob/profile", profileJSON, 409, checkIsErrorResponseJSON("Profile already exists. Use PUT.")},
		{"PUT", "/ob/profile", profileUpdateJSON, 200, checkIsJSON},
		{"PUT", "/ob/profile", profileUpdatedJSON, 200, checkIsJSON},
	})
}

func TestAvatar(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},
		{"POST", "/ob/avatar", avatarValidJSON, 200, checkIsEqualJSON(avatarValidJSONResponse)},
	})
}

func TestAvatarNoProfile(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"POST", "/ob/avatar", avatarValidJSON, 500, checkIsJSON},
	})
}

func TestAvatarUnexpectedEOF(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},
		{"POST", "/ob/avatar", avatarUnexpectedEOFJSON, 500, checkIsEqualJSON(avatarUnexpectedEOFJSONResponse)},
	})
}

func TestAvatarInvalidJSON(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},
		{"POST", "/ob/avatar", avatarInvalidJSON, 500, checkIsEqualJSON(avatarInvalidTQJSONResponse)},
	})
}

func TestImages(t *testing.T) {
	t.Parallel()

	// Valid image
	runAPITests(t, []apiTest{
		{"POST", "/ob/images", imageValidJSON, 200, checkIsEqualJSON(imageValidJSONResponse)},
	})
}

func TestHeader(t *testing.T) {
	t.Parallel()
	// It succeeds if we have a profile and the image data is valid
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},
		{"POST", "/ob/header", headerValidJSON, 200, checkIsEqualJSON(headerValidJSONResponse)},
	})
}

func TestHeaderNoProfile(t *testing.T) {
	t.Parallel()

	// Setting an header fails if we don't have a profile
	runAPITests(t, []apiTest{
		{"POST", "/ob/header", headerValidJSON, 500, checkIsJSON},
	})
}

func TestModerator(t *testing.T) {
	t.Parallel()

	// Fails without profile
	runAPITests(t, []apiTest{
		{"PUT", "/ob/moderator", moderatorValidJSON, http.StatusConflict, checkIsJSON},
	})

	// Works with profile
	runAPITests(t, []apiTest{
		{"POST", "/ob/profile", profileJSON, 200, checkIsJSON},

		// TODO: Enable after fixing bug that requires peers in order to set moderator status
		// {"PUT", "/ob/moderator", moderatorValidJSON, 200, checkIsEmptyJSONObject},

		// // Update
		// {"PUT", "/ob/moderator", moderatorUpdatedValidJSON, 200, checkIsEmptyJSONObject},
		{"DELETE", "/ob/moderator", "", 200, checkIsEmptyJSONObject},
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
		{"GET", "/ob/listings", "", 200, checkIsEmptyJSONArray},
		{"GET", "/ob/inventory", "", 200, checkIsEmptyJSONObject},

		// Invalid creates
		{"POST", "/ob/listing", `{`, 400, checkIsErrorResponseJSON("unexpected EOF")},

		{"GET", "/ob/listings", "", 200, checkIsEmptyJSONArray},
		{"GET", "/ob/inventory", "", 200, checkIsEmptyJSONObject},

		// TODO: Add support for improved JSON matching to since contracts
		// change each test run due to signatures

		// Create/Get
		{"GET", "/ob/listing/ron-swanson-tshirt", "", 404, checkIsEqualJSON("Listing not found.")},
		{"POST", "/ob/listing", goodListingJSON, 200, checkIsEqualJSON(`{"slug": "ron-swanson-tshirt"}`)},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", 200, checkIsJSON},
		{"POST", "/ob/listing", updatedListingJSON, 409, checkIsEqualJSON("Listing not found.")},

		// TODO: Add support for improved JSON matching to since contracts
		// change each test run due to signatures
		{"GET", "/ob/listings", "", 200, checkIsJSON},

		// TODO: This returns `inventoryJSONResponse` but slices are unordered
		// so they don't get considered equal. Figure out a way to fix that.
		{"GET", "/ob/inventory", "", 200, checkIsJSON},

		// Update inventory
		{"POST", "/ob/inventory", inventoryUpdateJSON, 200, checkIsEmptyJSONObject},

		// Update/Get Listing
		{"PUT", "/ob/listing", updatedListingJSON, 200, checkIsEmptyJSONObject},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", 200, checkIsJSON},

		// Delete/Get
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", 200, checkIsEmptyJSONObject},
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", 404, checkIsErrorResponseJSON("Listing not found.")},
		{"GET", "/ob/listing/ron-swanson-tshirt", "", 404, checkIsErrorResponseJSON("Listing not found.")},

		// Mutate non-existing listings
		{"PUT", "/ob/listing", updatedListingJSON, 404, checkIsErrorResponseJSON("Listing not found.")},
		{"DELETE", "/ob/listing/ron-swanson-tshirt", "", 404, checkIsErrorResponseJSON("Listing not found.")},
	})
}

func TestCryptoListings(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	updatedListing := *listing

	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
		{"GET", "/ob/listing/crypto", jsonFor(t, &updatedListing), 200, checkIsJSON},

		{"PUT", "/ob/listing", jsonFor(t, &updatedListing), 200, checkIsEmptyJSONObject},
		{"PUT", "/ob/listing", jsonFor(t, &updatedListing), 200, checkIsEmptyJSONObject},
		{"GET", "/ob/listing/crypto", jsonFor(t, &updatedListing), 200, checkIsJSON},

		{"DELETE", "/ob/listing/crypto", "", 200, checkIsEmptyJSONObject},
		{"DELETE", "/ob/listing/crypto", "", 404, checkIsErrorResponseJSON("Listing not found.")},
		{"GET", "/ob/listing/crypto", "", 404, checkIsErrorResponseJSON("Listing not found.")},
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
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
		{"GET", "/ob/listing/crypto", jsonFor(t, listing), 200, checkIsJSON},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMax + 0.001
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMax + 0.01
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(outOfRangeErr.Error())},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMin - 0.001
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
	})

	listing.Metadata.PriceModifier = core.PriceModifierMin - 1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(outOfRangeErr.Error())},
	})
}

func TestListingsQuantity(t *testing.T) {
	t.Parallel()

	listing := factory.NewListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
	})

	listing.Item.Skus[0].Quantity = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsJSON},
	})

	listing.Item.Skus[0].Quantity = -1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsJSON},
	})
}

func TestCryptoListingsQuantity(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsEqualJSON(`{"slug": "crypto"}`)},
	})

	listing.Item.Skus[0].Quantity = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrCryptocurrencySkuQuantityInvalid.Error())},
	})

	listing.Item.Skus[0].Quantity = -1
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrCryptocurrencySkuQuantityInvalid.Error())},
	})
}

func TestCryptoListingsNoCoinType(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	listing.Metadata.CoinType = ""

	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrCryptocurrencyListingCoinTypeRequired.Error())},
	})
}

func TestCryptoListingsCoinDivisibilityIncorrect(t *testing.T) {
	t.Parallel()

	listing := factory.NewCryptoListing("crypto")
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 200, checkIsJSON},
	})

	listing.Metadata.CoinDivisibility = 1e7
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrListingCoinDivisibilityIncorrect.Error())},
	})

	listing.Metadata.CoinDivisibility = 0
	runAPITests(t, []apiTest{
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrListingCoinDivisibilityIncorrect.Error())},
	})
}

func TestCryptoListingsIllegalFields(t *testing.T) {
	t.Parallel()

	runTest := func(listing *pb.Listing, err error) {
		runAPITests(t, []apiTest{
			{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(err.Error())},
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
		{"POST", "/ob/listing", jsonFor(t, listing), 500, checkIsErrorResponseJSON(core.ErrMarketPriceListingIllegalField("item.price").Error())},
	})
}

func TestStatus(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/ob/status", "", 400, checkIsJSON},
		{"GET", "/ob/status/QmYwAPJzv5CZsnA625s3Xf2nemtYgPpHdWEz79ojWnPbdG", "", 200, checkIsJSON},
	})
}

func TestWallet(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/wallet/address", "", 200, checkIsEqualJSON(walletAddressJSONResponse)},
		{"GET", "/wallet/balance", "", 200, checkIsEqualJSON(walletBalanceJSONResponse)},
		{"GET", "/wallet/mnemonic", "", 200, checkIsEqualJSON(walletMnemonicJSONResponse)},
		{"POST", "/wallet/spend", spendJSON, 400, checkIsEqualJSON(insuffientFundsJSON)},
	})
}

func TestConfig(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		// TODO: Need better JSON matching
		{"GET", "/ob/config", "", 200, checkIsJSON},
	})
}

func Test404(t *testing.T) {
	t.Parallel()

	// Test undefined endpoints
	runAPITests(t, []apiTest{
		{"GET", "/ob/a", "{}", 404, checkIsErrorResponseJSON("Not Found")},
		{"PUT", "/ob/a", "{}", 404, checkIsErrorResponseJSON("Not Found")},
		{"POST", "/ob/a", "{}", 404, checkIsErrorResponseJSON("Not Found")},
		{"PATCH", "/ob/a", "{}", 404, checkIsErrorResponseJSON("Not Found")},
		{"DELETE", "/ob/a", "{}", 404, checkIsErrorResponseJSON("Not Found")},
	})
}

func TestPosts(t *testing.T) {
	t.Parallel()

	runAPITests(t, []apiTest{
		{"GET", "/ob/posts", "", 200, checkIsEmptyJSONArray},

		// Invalid creates
		{"POST", "/ob/post", `{`, 400, checkIsErrorResponseJSON("unexpected EOF")},

		{"GET", "/ob/posts", "", 200, checkIsEmptyJSONArray},

		// Create/Get
		{"GET", "/ob/post/test1", "", 404, checkIsErrorResponseJSON("Post not found.")},
		{"POST", "/ob/post", postJSON, 200, checkIsEqualJSON(postJSONResponse)},
		{"GET", "/ob/post/test1", "", 200, checkIsJSON},
		{"POST", "/ob/post", postUpdateJSON, 409, checkIsErrorResponseJSON("Post already exists. Use PUT.")},

		{"GET", "/ob/posts", "", 200, checkIsJSON},

		// Update/Get Post
		{"PUT", "/ob/post", postUpdateJSON, 200, checkIsEmptyJSONObject},
		{"GET", "/ob/post/test1", "", 200, checkIsJSON},

		// Delete/Get
		{"DELETE", "/ob/post/test1", "", 200, checkIsEmptyJSONObject},
		{"DELETE", "/ob/post/test1", "", 404, checkIsErrorResponseJSON("Post not found.")},
		{"GET", "/ob/post/test1", "", 404, checkIsErrorResponseJSON("Post not found.")},

		// Mutate non-existing listings
		{"PUT", "/ob/post", postUpdateJSON, 404, checkIsErrorResponseJSON("Post not found.")},
		{"DELETE", "/ob/post/test1", "", 404, checkIsErrorResponseJSON("Post not found.")},
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
		{"POST", "/ob/closedispute", expiredPostJSON, 400, checkIsJSON},
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
		{"POST", "/ob/releaseescrow", fmt.Sprintf(`{"orderId":"%s"}`, sale.OrderID), 400, checkIsJSON},
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

	checkNewSale := func(t *testing.T, respBody []byte) {
		respObj := struct {
			Sales []repo.Sale `json:"sales"`
		}{}
		err := json.Unmarshal(respBody, &respObj)
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
		{"GET", "/ob/sales", "", 200, checkNewSale},
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

	checkNewPurchase := func(t *testing.T, respBody []byte) {
		respObj := struct {
			Purchases []repo.Purchase `json:"purchases"`
		}{}
		err := json.Unmarshal(respBody, &respObj)
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
		{"GET", "/ob/purchases", "", 200, checkNewPurchase},
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

	checkNewCase := func(t *testing.T, respBody []byte) {
		respObj := struct {
			Cases []repo.Case `json:"cases"`
		}{}
		err := json.Unmarshal(respBody, &respObj)
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
		{"GET", "/ob/cases", "", 200, checkNewCase},
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
		{"GET", "/ob/notifications?limit=-1", "", 200, checkIsEqualJSON(sameTimestampsAreReturnedInReverse)},
		{"GET", "/ob/notifications?limit=-1&offsetId=notif3", "", 200, checkIsEqualJSON(sameTimestampsAreReturnedInReverseAndRespectOffsetID)},
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
//{"POST", "/ob/profile", moderatorProfileJSON, 200, checkIsJSON},
//{"POST", "/ob/closedispute", nonexpiredPostJSON, 200, checkIsJSON},
//}, dbSetup, nil)
//}
