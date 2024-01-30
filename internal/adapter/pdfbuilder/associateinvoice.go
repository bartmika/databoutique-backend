package pdfbuilder

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/leekchan/accounting"
	"github.com/signintech/gopdf"
	"go.mongodb.org/mongo-driver/bson/primitive"

	c "github.com/bartmika/databoutique-backend/internal/config"
)

type PDFBuilderResponseDTO struct {
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
	Content  []byte `json:"content"`
}

type AssociateInvoiceBuilderRequestDTO struct {
	// Country controls what currency formatting we will apply to the invoice.
	// If nothing is set then we will automatically assume `Canada` and apply
	// the formatting for this country.
	Country string `bson:"country" json:"country"`

	// accounting variable to store our currency formatter. Keep this unexported
	// because we do not want to store it.
	accounting accounting.Accounting `bson:"-" json:"-"`

	ID                       primitive.ObjectID `json:"id"`
	TenantID                 primitive.ObjectID `json:"tenant_id"`
	OrderID                  primitive.ObjectID `json:"order_id"`
	OrderWJID                uint64             `json:"order_wjid"`
	InvoiceID                string             `json:"invoice_id"`
	InvoiceDate              time.Time          `json:"invoice_date"`
	AssociateName            string             `json:"associate_name"`
	AssociatePhone           string             `json:"associate_phone"`
	ClientName               string             `json:"client_name"`
	ClientPhone              string             `json:"client_phone"`
	ClientEmail              string             `json:"client_email"`
	Line01Qty                int64              `json:"line_01_qty"`
	Line01Desc               string             `json:"line_01_desc"`
	Line01Price              float64            `json:"line_01_price"`
	Line01Amount             float64            `json:"line_01_amount"`
	Line02Qty                int64              `json:"line_02_qty"`
	Line02Desc               string             `json:"line_02_desc"`
	Line02Price              float64            `json:"line_02_price"`
	Line02Amount             float64            `json:"line_02_amount"`
	Line03Qty                int64              `json:"line_03_qty"`
	Line03Desc               string             `json:"line_03_desc"`
	Line03Price              float64            `json:"line_03_price"`
	Line03Amount             float64            `json:"line_03_amount"`
	Line04Qty                int64              `json:"line_04_qty"`
	Line04Desc               string             `json:"line_04_desc"`
	Line04Price              float64            `json:"line_04_price"`
	Line04Amount             float64            `json:"line_04_amount"`
	Line05Qty                int64              `json:"line_05_qty"`
	Line05Desc               string             `json:"line_05_desc"`
	Line05Price              float64            `json:"line_05_price"`
	Line05Amount             float64            `json:"line_05_amount"`
	Line06Qty                int64              `json:"line_06_qty"`
	Line06Desc               string             `json:"line_06_desc"`
	Line06Price              float64            `json:"line_06_price"`
	Line06Amount             float64            `json:"line_06_amount"`
	Line07Qty                int64              `json:"line_07_qty"`
	Line07Desc               string             `json:"line_07_desc"`
	Line07Price              float64            `json:"line_07_price"`
	Line07Amount             float64            `json:"line_07_amount"`
	Line08Qty                int64              `json:"line_08_qty"`
	Line08Desc               string             `json:"line_08_desc"`
	Line08Price              float64            `json:"line_08_price"`
	Line08Amount             float64            `json:"line_08_amount"`
	Line09Qty                int64              `json:"line_09_qty"`
	Line09Desc               string             `json:"line_09_desc"`
	Line09Price              float64            `json:"line_09_price"`
	Line09Amount             float64            `json:"line_09_amount"`
	Line10Qty                int64              `json:"line_10_qty"`
	Line10Desc               string             `json:"line_10_desc"`
	Line10Price              float64            `json:"line_10_price"`
	Line10Amount             float64            `json:"line_10_amount"`
	Line11Qty                int64              `json:"line_11_qty"`
	Line11Desc               string             `json:"line_11_desc"`
	Line11Price              float64            `json:"line_11_price"`
	Line11Amount             float64            `json:"line_11_amount"`
	Line12Qty                int64              `json:"line_12_qty"`
	Line12Desc               string             `json:"line_12_desc"`
	Line12Price              float64            `json:"line_12_price"`
	Line12Amount             float64            `json:"line_12_amount"`
	Line13Qty                int64              `json:"line_13_qty"`
	Line13Desc               string             `json:"line_13_desc"`
	Line13Price              float64            `json:"line_13_price"`
	Line13Amount             float64            `json:"line_13_amount"`
	Line14Qty                int64              `json:"line_14_qty"`
	Line14Desc               string             `json:"line_14_desc"`
	Line14Price              float64            `json:"line_14_price"`
	Line14Amount             float64            `json:"line_14_amount"`
	Line15Qty                int64              `json:"line_15_qty"`
	Line15Desc               string             `json:"line_15_desc"`
	Line15Price              float64            `json:"line_15_price"`
	Line15Amount             float64            `json:"line_15_amount"`
	InvoiceQuoteDays         int8               `json:"invoice_quote_days"`
	InvoiceAssociateTax      string             `json:"invoice_associate_tax"`
	InvoiceQuoteDate         time.Time          `json:"invoice_quote_date"`
	InvoiceCustomersApproval string             `json:"invoice_customers_approval"`
	Line01Notes              string             `json:"line_01_notes"`
	Line02Notes              string             `json:"line_02_notes"`
	TotalLabour              float64            `json:"total_labour"`
	TotalMaterials           float64            `json:"total_materials"`
	OtherCosts               float64            `json:"other_costs"`
	SubTotal                 float64            `json:"sub_total"`
	Tax                      float64            `json:"tax"`
	Total                    float64            `json:"total"`
	PaymentAmount            float64            `json:"payment_amount"`
	PaymentDate              time.Time          `json:"payment_date"`
	IsCash                   bool               `json:"is_cash"`
	IsCheque                 bool               `json:"is_cheque"`
	IsDebit                  bool               `json:"is_debit"`
	IsCredit                 bool               `json:"is_credit"`
	IsOther                  bool               `json:"is_other"`
	ClientSignature          string             `json:"client_signature"`
	AssociateSignDate        time.Time          `json:"associate_sign_date"`
	AssociateSignature       string             `json:"associate_signature"`
	CreatedTime              string             `json:"created_time"`
	LastModifiedTime         string             `json:"last_modified_time"`
	CreatedById              primitive.ObjectID `json:"created_by_id"`
	CreatedByName            string             `json:"created_by_name"`
	LastModifiedById         primitive.ObjectID `json:"last_modified_by_id"`
	LastModifiedByName       string             `json:"last_modified_by_name"`
	CreatedFrom              string             `json:"created_from"`
	CreatedFromIsPublic      string             `json:"created_from_is_public"`
	LastModifiedFrom         string             `json:"last_modified_from"`
	LastModifiedFromIsPublic string             `json:"last_modified_from_is_public"`
	ClientAddress            string             `json:"client_address"`
	Version                  uint64             `json:"version"`
	Deposit                  float64            `json:"deposit"`
	AmountDue                float64            `json:"amount_due"`
	Status                   string             `json:"state"` // IsArchived string `json:"is_archived"`
}

type AssociateInvoiceBuilder interface {
	GeneratePDF(dto *AssociateInvoiceBuilderRequestDTO) (*PDFBuilderResponseDTO, error)
}

type aiBuilder struct {
	PDFTemplateFilePath string
	DataDirectoryPath   string
	Logger              *slog.Logger
}

func NewAssociateInvoiceBuilder(cfg *c.Conf, logger *slog.Logger) AssociateInvoiceBuilder {
	// Defensive code: Make sure we have access to the file before proceeding any further with the code.
	logger.Debug("pdf builder for ai initializing...")
	_, err := os.Stat(cfg.PDFBuilder.AssociateInvoiceTemplatePath)
	if os.IsNotExist(err) {
		log.Fatal(errors.New("file does not exist"))
	}

	return &aiBuilder{
		PDFTemplateFilePath: cfg.PDFBuilder.AssociateInvoiceTemplatePath,
		DataDirectoryPath:   cfg.PDFBuilder.DataDirectoryPath,
		Logger:              logger,
	}
}

func (bdr *aiBuilder) GeneratePDF(r *AssociateInvoiceBuilderRequestDTO) (*PDFBuilderResponseDTO, error) {
	var err error
	bdr.Logger.Debug("opening up template file", slog.String("file", bdr.PDFTemplateFilePath))

	// Create our currency formatter. Techdebt for adding future formats.
	switch r.Country {
	// DEVELOPERS NOTE: This is techdebt - add support for future countries here.
	default:
		// See: https://github.com/leekchan/accounting
		r.accounting = accounting.Accounting{Symbol: "$", Precision: 2}
	}

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: gopdf.Unit_PT, PageSize: gopdf.Rect{W: 595.28, H: 841.89}}) //595.28, 841.89 = A4
	pdf.AddPage()

	// DEVELOPER NOTE:
	// The `github.com/signintech/gopdf` library needs to access a `tff` file
	// to utilize to render font family in our PDF. Therefore the following set
	// of lines are going to populate the font family we will need to use,
	err = pdf.AddTTFFont("roboto", "./static/roboto/Roboto-Regular.ttf")
	if err != nil {
		panic(err) // For developer purpose only.
	}
	err = pdf.AddTTFFont("roboto-bold", "./static/roboto/Roboto-Bold.ttf")
	if err != nil {
		panic(err) // For developer purpose only.
	}

	//// Import page 1
	tpl1 := pdf.ImportPage(bdr.PDFTemplateFilePath, 1, "/MediaBox")
	pdf.UseImportedTemplate(tpl1, 0, 0, 595.28, 841.89) // Draw imported template onto page

	//// Generate page one contents.
	if err := bdr.generateContent(r, &pdf); err != nil {
		return nil, err
	}

	////
	//// Generate the file and save it to the file.
	////

	fileName := fmt.Sprintf("invoice_%d_v%d.pdf", r.OrderWJID, r.Version)
	filePath := fmt.Sprintf("%s/%s", bdr.DataDirectoryPath, fileName)

	err = pdf.WritePdf(filePath)
	if err != nil {
		return nil, err
	}

	////
	//// Open the file and read all the binary data.
	////

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bin, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	/////

	// cleanup(bdr.DataDirectoryPath)

	////
	//// Return the generate invoice.
	////

	return &PDFBuilderResponseDTO{
		FileName: fileName,
		FilePath: filePath,
		Content:  bin,
	}, err
}

func (bdr *aiBuilder) generateContent(dto *AssociateInvoiceBuilderRequestDTO, pdf *gopdf.GoPdf) error {
	// Invoice Date
	pdf.SetFont("roboto", "", 12)
	pdf.SetXY(400, 71) // x coordinate specification

	// Associate
	pdf.Cell(nil, dto.InvoiceDate.Format("2006-01-02"))
	pdf.SetXY(410, 93)
	pdf.Cell(nil, dto.AssociateName)
	pdf.SetXY(410, 115)
	pdf.Cell(nil, dto.AssociatePhone)

	// Client
	pdf.SetXY(135, 138)
	pdf.Cell(nil, dto.ClientName)
	pdf.SetXY(115, 160)
	pdf.Cell(nil, dto.ClientAddress)
	pdf.SetXY(110, 182)
	pdf.Cell(nil, dto.ClientPhone)
	pdf.SetXY(270, 182)
	pdf.Cell(nil, dto.ClientEmail)

	// List items.
	// --- Line 1 ---
	if dto.Line01Qty == 0 {
		return fmt.Errorf("`line_01_qty` is missing or cannot be zero: %v", dto.Line01Qty)
	}
	if dto.Line01Desc == "" {
		return fmt.Errorf("`line_01_desc` is missing or cannot be empty: %v", dto.Line01Desc)
	}
	if dto.Line01Price == 0 {
		return fmt.Errorf("`line_01_price` is missing or cannot be zero: %v", dto.Line01Price)
	}
	if dto.Line01Amount == 0 {
		return fmt.Errorf("`line_01_amount` is missing or cannot be zero: %v", dto.Line01Amount)
	}
	pdf.SetXY(87, 219)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line01Qty))
	pdf.SetXY(140, 219)
	pdf.Cell(nil, dto.Line01Desc)
	pdf.SetXY(390, 219)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line01Price))
	pdf.SetXY(490, 219)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line01Amount))
	// --- Line 2 ---
	pdf.SetXY(87, 235) // +16
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line02Qty))
	pdf.SetXY(140, 235)
	pdf.Cell(nil, dto.Line02Desc)
	pdf.SetXY(390, 235)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line02Price))
	pdf.SetXY(490, 235)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line02Amount))
	// --- Line 3 ---
	pdf.SetXY(87, 251)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line03Qty))
	pdf.SetXY(140, 251)
	pdf.Cell(nil, dto.Line03Desc)
	pdf.SetXY(390, 251)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line03Price))
	pdf.SetXY(490, 251)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line03Amount))
	// --- Line 4 ---
	pdf.SetXY(87, 267)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line04Qty))
	pdf.SetXY(140, 267)
	pdf.Cell(nil, dto.Line04Desc)
	pdf.SetXY(390, 267)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line04Price))
	pdf.SetXY(490, 267)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line04Amount))
	// --- Line 5 ---
	pdf.SetXY(87, 283)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line05Qty))
	pdf.SetXY(140, 283)
	pdf.Cell(nil, dto.Line05Desc)
	pdf.SetXY(390, 283)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line05Price))
	pdf.SetXY(490, 283)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line05Amount))
	// --- Line 6 ---
	pdf.SetXY(87, 299)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line06Qty))
	pdf.SetXY(140, 299)
	pdf.Cell(nil, dto.Line06Desc)
	pdf.SetXY(390, 299)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line06Price))
	pdf.SetXY(490, 299)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line06Amount))
	// --- Line 7 ---
	pdf.SetXY(87, 315)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line07Qty))
	pdf.SetXY(140, 315)
	pdf.Cell(nil, dto.Line07Desc)
	pdf.SetXY(390, 315)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line07Price))
	pdf.SetXY(490, 315)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line07Amount))
	// --- Line 8 ---
	pdf.SetXY(87, 331)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line08Qty))
	pdf.SetXY(140, 331)
	pdf.Cell(nil, dto.Line08Desc)
	pdf.SetXY(390, 331)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line08Price))
	pdf.SetXY(490, 331)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line08Amount))
	// --- Line 9 ---
	pdf.SetXY(87, 347)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line09Qty))
	pdf.SetXY(140, 347)
	pdf.Cell(nil, dto.Line09Desc)
	pdf.SetXY(390, 347)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line09Price))
	pdf.SetXY(490, 347)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line09Amount))
	// --- Line 10 ---
	pdf.SetXY(87, 363)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line10Qty))
	pdf.SetXY(140, 363)
	pdf.Cell(nil, dto.Line10Desc)
	pdf.SetXY(390, 363)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line10Price))
	pdf.SetXY(490, 363)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line10Amount))
	// --- Line 11 ---
	pdf.SetXY(87, 379)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line11Qty))
	pdf.SetXY(140, 379)
	pdf.Cell(nil, dto.Line11Desc)
	pdf.SetXY(390, 379)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line11Price))
	pdf.SetXY(490, 379)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line11Amount))
	// --- Line 12 ---
	pdf.SetXY(87, 395)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line12Qty))
	pdf.SetXY(140, 395)
	pdf.Cell(nil, dto.Line12Desc)
	pdf.SetXY(390, 395)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line12Price))
	pdf.SetXY(490, 395)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line12Amount))
	// --- Line 13 ---
	pdf.SetXY(87, 411)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line13Qty))
	pdf.SetXY(140, 411)
	pdf.Cell(nil, dto.Line13Desc)
	pdf.SetXY(390, 411)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line13Price))
	pdf.SetXY(490, 411)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line13Amount))
	// --- Line 14 ---
	pdf.SetXY(87, 427)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line14Qty))
	pdf.SetXY(140, 427)
	pdf.Cell(nil, dto.Line14Desc)
	pdf.SetXY(390, 427)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line14Price))
	pdf.SetXY(490, 427)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line14Amount))
	// --- Line 15 ---
	pdf.SetXY(87, 443)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.Line15Qty))
	pdf.SetXY(140, 443)
	pdf.Cell(nil, dto.Line15Desc)
	pdf.SetXY(390, 443)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line15Price))
	pdf.SetXY(490, 443)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Line15Amount))

	// Valid for X amount of days.
	pdf.SetXY(190, 475)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.InvoiceQuoteDays))

	// Associate HST # if applicable
	pdf.SetXY(223, 523)
	pdf.Cell(nil, dto.InvoiceAssociateTax)
	pdf.SetXY(195, 550)
	pdf.Cell(nil, dto.InvoiceQuoteDate.Format("2006-01-02"))
	pdf.SetXY(210, 575)
	pdf.Cell(nil, dto.InvoiceCustomersApproval)

	// Totals
	pdf.SetXY(465, 465)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.TotalLabour))
	pdf.SetXY(465, 491)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.TotalMaterials))
	pdf.SetXY(465, 513)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.OtherCosts))
	pdf.SetXY(465, 539)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.SubTotal))
	pdf.SetXY(465, 559)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Tax))
	pdf.SetXY(465, 590)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Total))

	// Notes
	pdf.SetXY(60, 633)
	pdf.Cell(nil, dto.Line01Notes)
	pdf.SetXY(60, 665)
	pdf.Cell(nil, dto.Line02Notes)

	// Deposit
	pdf.SetXY(460, 669)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.Deposit))

	// Payment Date
	pdf.SetXY(160, 693)
	pdf.Cell(nil, dto.PaymentDate.Format("2006-01-02"))

	// Payment Amount Due
	pdf.SetXY(460, 692)
	pdf.Cell(nil, dto.accounting.FormatMoney(dto.PaymentAmount))

	// Payment Methods
	pdf.SetXY(77, 714)
	if dto.IsCash {
		pdf.Cell(nil, "Y")
	} else {
		pdf.Cell(nil, "N")
	}

	pdf.SetXY(177, 714)
	if dto.IsCheque {
		pdf.Cell(nil, "Y")
	} else {
		pdf.Cell(nil, "N")
	}

	pdf.SetXY(277, 714)
	if dto.IsDebit {
		pdf.Cell(nil, "Y")
	} else {
		pdf.Cell(nil, "N")
	}

	pdf.SetXY(377, 714)
	if dto.IsCredit {
		pdf.Cell(nil, "Y")
	} else {
		pdf.Cell(nil, "N")
	}

	pdf.SetXY(477, 714)
	if dto.IsOther {
		pdf.Cell(nil, "Y")
	} else {
		pdf.Cell(nil, "N")
	}

	// Client signature upon completion
	pdf.SetXY(260, 735)
	pdf.Cell(nil, dto.ClientSignature)

	// Associate signed on
	pdf.SetXY(120, 762)
	pdf.Cell(nil, dto.AssociateSignDate.Format("2006-01-02"))
	pdf.SetXY(380, 762)
	pdf.Cell(nil, dto.AssociateSignature)

	// Workery Job ID #
	pdf.SetXY(469, 805)
	pdf.Cell(nil, fmt.Sprintf("%v", dto.OrderWJID))

	return nil
}
