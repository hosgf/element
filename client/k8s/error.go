package k8s

import "github.com/gogf/gf/v2/text/gstr"

func ToError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) < 1 {
		return msg
	}
	if gstr.ContainsI(msg, "namespaces \"driver\" not found") {

	}
	return msg
}
