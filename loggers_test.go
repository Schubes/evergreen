package evergreen

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/tychoish/grip/send"
	"github.com/tychoish/grip/slogger"
)

func TestLoggingWriter(t *testing.T) {
	Convey("With a Logger", t, func() {
		Convey("data written to Writer should match data appended", func() {
			sender := send.MakeInternalLogger()
			testLogger := &slogger.Logger{
				Name:      "",
				Appenders: []send.Sender{sender},
			}
			logWriter := NewInfoLoggingWriter(testLogger)

			testLogLine := "blah blah %x %fkjabsddg"
			logWriter.Write([]byte(testLogLine + "\n"))
			So(sender.Len(), ShouldEqual, 1)
			So(strings.HasSuffix(sender.GetMessage().Rendered, testLogLine), ShouldBeTrue)
		})

		Convey("writer should cache data until a new line is sent", func() {
			sender := send.MakeInternalLogger()
			testLogger := &slogger.Logger{
				Name:      "",
				Appenders: []send.Sender{sender},
			}
			logWriter := NewInfoLoggingWriter(testLogger)

			msgs := []string{"foo bar", "bar", "foo", "baz baz baz!!!\n"}
			for _, msg := range msgs {
				logWriter.Write([]byte(msg))
			}
			So(sender.GetMessage().Rendered, ShouldEndWith, strings.Trim(strings.Join(msgs, ""), "\n"))
		})
	})
}
