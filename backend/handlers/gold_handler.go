package handlers

import (
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"regexp"
	"time"
	"wealth-management/domains"
	"wealth-management/repository"

	"github.com/cockroachdb/apd/v3"
	"github.com/gin-gonic/gin"
)

type GoldHandler struct {
	fundRepo repository.GoldRepository
}

// NewGoldHandler it might be fine without but it is a contructor so fundRepo stay private and unmodified when created.
func NewGoldHandler(fundRepo repository.GoldRepository) GoldHandler {
	return GoldHandler{fundRepo: fundRepo}
}

func (handler *GoldHandler) GetAllGoldsTxn(context *gin.Context) {
	funds, err := handler.fundRepo.GetAll()
	if err != nil {
		log.Printf("Error getting funds: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	context.JSON(http.StatusOK, gin.H{"golds": funds})
}

// todo refactor after function test ok
// todo remove all entry that from "file" before insert
// PostBulkImportGolds expect client send files with multipart/form-data
func (handler *GoldHandler) PostBulkImportGolds(context *gin.Context) {
	fileHeader, err := context.FormFile("file")
	if err != nil {
		log.Printf("Error getting file header: %v", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer file.Close()
	csvReader := csv.NewReader(file)

	// recognize header position
	// at the moment support: "Investment Type,Bank,Investment Date	Selling Date	Gold (Gram)	Purchase Unit Price	Selling Unit Price	Gainc/ Loss	Status	Remarks"
	headerRow, err := csvReader.Read()
	if err != nil {
		log.Printf("Error reading header: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	indexByHeaderMap := make(map[string]int)
	for index, header := range headerRow {
		var key string
		if header == "Bank" {
			key = "Bank"
		} else if header == "Purchase Unit Price" {
			key = "UnitPrice"
		} else if header == "Investment Date" {
			key = "TxnDate"
		} else if header == "Gold (Gram)" {
			key = "Gram"
		} else if header == "Status" {
			key = "TxnType"
		}
		if key != "" {
			indexByHeaderMap[key] = index
		}
	}

	var goldTxns []domains.GoldTxn
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading csv: %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// todo do the mapping logic here base on the header recognize logic
		// golang date format design is just nut, the layout MUST follow their references
		dateStr := record[indexByHeaderMap["TxnDate"]]
		if dateStr == "" {
			log.Println("No TxnDate, Dirty Data found, skipping this row")
			continue
		}
		txnDate, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			log.Printf("Error parsing txnDate: %v, source: %s", err, dateStr)
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		txnType := "BUY"
		if record[indexByHeaderMap["TxnType"]] == "Sold Out" {
			txnType = "SOLD"
		}

		gram, _, err := apd.NewFromString(toNumericString(record[indexByHeaderMap["Gram"]]))
		if err != nil {
			log.Printf("Error parsing gold txn: %v, source string %s", err, record[indexByHeaderMap["Gram"]])
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		unitPrice, _, err := apd.NewFromString(toNumericString(record[indexByHeaderMap["UnitPrice"]]))
		if err != nil {
			log.Printf("Error parsing gold unit price: %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx := apd.Context{
			Precision: 12,
		}

		totalPrice := new(apd.Decimal)
		_, err = ctx.Mul(totalPrice, gram, unitPrice)
		if err != nil {
			log.Printf("Error creating condition: %v", err)
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Txn Date %s\n", record[indexByHeaderMap["TxnDate"]])

		goldTxns = append(goldTxns, domains.GoldTxn{
			Bank:        record[indexByHeaderMap["Bank"]],
			TxnDate:     txnDate,
			Gram:        *gram,
			UnitPrice:   *unitPrice,
			TotalPrice:  *totalPrice,
			TxnType:     txnType,
			EntrySource: "BulkImport",
		})
	}

	log.Printf("Debug: %v", goldTxns)
	// todo: remove all rows where EntrySource: "BulkImport" and insert back in

	context.JSON(http.StatusOK, gin.H{"message": "Gold file import successful"})
}

func toNumericString(str string) string {
	return regexp.MustCompile("[^0-9.]").ReplaceAllString(str, "")

}
