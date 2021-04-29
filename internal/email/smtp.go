package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// MaxLineLength is the maximum line length per RFC 2045
	MaxLineLength = 76
	delimeter     = "**=myohmy689407924327898338383"
)

type SmtpConfig struct {
	Server       string
	Username     string
	Password     string
	FromOverride string
	Helo         string
	InsecureTLS  bool
}

// Builder builds emails
type Builder struct {
	From    string
	To      string
	ReplyTo string
	Body    string
	Subject string

	attachments []emailAttachment
}

type emailAttachment struct {
	filename    string
	contentType string
	data        io.Reader
}

func sanitizeAttachmentName(name string) string {
	return filepath.Base(name)
}

// trimAddresses workaround for go < 1.15
func trimAddresses(address string) string {
	return strings.Trim(strings.Trim(address, " "), ",")
}

// AddFile adds a file attachment
func (b *Builder) AddFile(name string, data io.Reader, contentType string) {
	log.Debugln("Adding file: ", name, " contentType: ", contentType)
	if contentType == "" {
		log.Warnln("no contentType, setting to binary")
		contentType = "application/octet-stream"
	}
	attachment := emailAttachment{
		contentType: contentType,
		filename:    sanitizeAttachmentName(name),
		data:        data,
	}
	b.attachments = append(b.attachments, attachment)
}

func (b *Builder) WriteAttachments(w io.Writer) (err error) {
	for _, attachment := range b.attachments {
		log.Debugln("File attachment: ", attachment.filename)

		fileHeader := fmt.Sprintf("\r\n--%s\r\n", delimeter)
		fileHeader += "Content-Type: " + attachment.contentType + "; charset=\"utf-8\"\r\n"
		fileHeader += "Content-Transfer-Encoding: base64\r\n"
		fileHeader += "Content-Disposition: attachment;filename=\"" + attachment.filename + "\"\r\n\r\n"
		_, err = w.Write([]byte(fileHeader))
		if err != nil {
			return err
		}

		splittingEncoder := &SplittingWritter{
			innerWriter:    w,
			maxLineLength:  MaxLineLength,
			lineTerminator: "\r\n",
		}
		base64Encoder := base64.NewEncoder(base64.StdEncoding, splittingEncoder)
		_, err := io.Copy(base64Encoder, attachment.data)

		if err != nil {
			return err
		}
		base64Encoder.Close()
	}
	return nil
}

// Send sends the email
func (b *Builder) Send(cfg *SmtpConfig) (err error) {
	if cfg == nil {
		return fmt.Errorf("no smtp config")
	}
	frm := b.From
	if cfg.FromOverride != "" {
		frm = cfg.FromOverride
	}
	//if not defined
	from, err := mail.ParseAddress(frm)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddressList(trimAddresses(b.To))
	if err != nil {
		return err
	}

	log.Debug("from:", from)
	log.Debug("to:", to)

	host, _, _ := net.SplitHostPort(cfg.Server)

	tlsconfig := &tls.Config{
		InsecureSkipVerify: cfg.InsecureTLS,
		ServerName:         host,
	}

	conn, err := tls.Dial("tcp", cfg.Server, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	if cfg.Helo != "" {
		err = c.Hello(cfg.Helo)
		if err != nil {
			return err
		}
	}

	if cfg.Username != "" {
		auth := smtp.PlainAuth("", cfg.Username, cfg.Password, host)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	if err = c.Mail(from.Address); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr.Address); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}
	//basic email headers
	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", b.To)
	msg += fmt.Sprintf("Subject: %s\r\n", b.Subject)
	// msg += fmt.Sprintf("ReplyTo: %s\r\n", b.ReplyTo)

	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)

	msg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	msg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	msg += "Content-Transfer-Encoding: quoted-printable\r\n"
	msg += "Content-Disposition: inline\r\n"
	msg += "\r\n"
	msg += b.Body

	log.Debug("mime msg", msg)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}

	err = b.WriteAttachments(w)
	if err != nil {
		return err
	}

	// Add last boundary delimeter, with trailing -- according to RFC 1341
	lastBoundary := fmt.Sprintf("\r\n--%s--\r\n", delimeter)
	_, err = w.Write([]byte(lastBoundary))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Info("Message sent")
	return nil
}

// SplittingWritter writes a stream and inserts a terminar
type SplittingWritter struct {
	innerWriter       io.Writer
	currentLineLength int
	maxLineLength     int
	lineTerminator    string
}

func (w *SplittingWritter) Write(p []byte) (n int, err error) {
	length := len(p)
	total := 0
	for to, from := 0, 0; from < length; from = to {
		delta := w.maxLineLength - w.currentLineLength

		to = from + delta
		if to > length {
			to = length
			delta = length - from
		}

		n, err = w.innerWriter.Write(p[from:to])
		total += n
		if err != nil {
			return total, err
		}

		w.currentLineLength += delta

		if w.currentLineLength == w.maxLineLength {
			n, err = w.innerWriter.Write([]byte(w.lineTerminator))
			total += n
			if err != nil {
				return total, err
			}
			w.currentLineLength = 0
		}
	}

	return total, nil
}

func chunkSplit(body string, limit int, end string) string {
	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	var result = ""

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit
		result = result + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		if len(charSlice) < limit {
			limit = len(charSlice)
		}
	}
	return result
}
