package handlers

import (
	"encoding/csv"
	"log"
	"net/http"
	"regexp"
	"time"
	"wealth-management/domains"
	"wealth-management/repository"

	"github.com/cockroachdb/apd/v3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoldHandler struct {
	goldRepo repository.GoldRepository
}

// NewGoldHandler it might be fine without but it is a contructor so goldRepo stay private and unmodified when created.
func NewGoldHandler(fundRepo repository.GoldRepository) GoldHandler {
	return GoldHandler{goldRepo: fundRepo}
}

func (handler *GoldHandler) GetAllGoldsTxn(context *gin.Context) {
	funds, err := handler.goldRepo.GetAll()
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
		return
	}
	defer file.Close()
	csvReader := csv.NewReader(file)

	// recognize header position
	// at the moment support: "Investment Type,Bank,Investment Date	Selling Date	Gold (Gram)	Purchase Unit Price	Selling Unit Price	Gainc/ Loss	Status	Remarks"
	indexByHeaderMap, err := identifyHeader(csvReader)
	if err != nil {
		log.Printf("Error reading file header: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// read remaining rows
	rows, err := csvReader.ReadAll()
	if err != nil {
		log.Printf("Error reading file content: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	goldTxns, err := parseGoldTxns(rows, indexByHeaderMap)
	if err != nil {
		log.Printf("Error parsing content: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = handler.goldRepo.ReplaceAllByEntrySource("BulkImport", goldTxns)
	if err != nil {
		log.Printf("Error Replacing By Entry Source: %v", err)
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, gin.H{"message": "Gold file import successful"})
}

func parseGoldTxns(rows [][]string, indexByHeaderMap map[string]int) ([]domains.GoldTxn, error) {
	var goldTxns []domains.GoldTxn

	for _, record := range rows {
		dateStr := record[indexByHeaderMap["TxnDate"]]
		if dateStr == "" {
			log.Println("No TxnDate, consider as dirty data, skipping this row")
			continue
		}
		// golang date format design is just nut, the layout MUST follow their references example value
		txnDate, err := time.Parse("January 2, 2006", dateStr)
		if err != nil {
			return nil, err
		}

		txnType := "BUY"
		if record[indexByHeaderMap["TxnType"]] == "Sold Out" {
			txnType = "SOLD"
		}

		gram, err := toDecimal(record[indexByHeaderMap["Gram"]])
		if err != nil {
			return nil, err
		}
		unitPrice, err := toDecimal(record[indexByHeaderMap["UnitPrice"]])
		if err != nil {
			return nil, err
		}

		ctx := apd.BaseContext.WithPrecision(12)
		totalPrice := new(apd.Decimal)
		_, err = ctx.Mul(totalPrice, gram, unitPrice)
		if err != nil {
			return nil, err
		}

		goldTxns = append(goldTxns, domains.GoldTxn{
			ID:          uuid.New().String(),
			Bank:        record[indexByHeaderMap["Bank"]],
			TxnDate:     txnDate,
			Gram:        *gram,
			UnitPrice:   *unitPrice,
			TotalPrice:  *totalPrice,
			TxnType:     txnType,
			EntrySource: "BulkImport",
		})
	}
	return goldTxns, nil
}

func identifyHeader(csvReader *csv.Reader) (map[string]int, error) {
	headerRow, err := csvReader.Read()
	if err != nil {
		return nil, err
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
	return indexByHeaderMap, nil
}

func toDecimal(str string) (*apd.Decimal, error) {
	numericStr := regexp.MustCompile("[^0-9.]").ReplaceAllString(str, "")
	numeric, _, err := apd.NewFromString(numericStr)
	if err != nil {
		return nil, err
	}
	return numeric, nil
}
