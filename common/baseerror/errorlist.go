package baseerror

var (
	ErrSystem       = NewCodeError(10001, "服务器错误")
	ErrInvalidParam = NewCodeError(10002, "参数错误")
	ErrRequireLogin = NewCodeError(10001, "登陆密码已修改或失效，请重新登陆")
	ErrUserDisabled = NewCodeError(10001, "检测到您的账号存在发布不良信息等问题，现已被封号禁止使用，如果您有疑问，可联系晓黑板客服！")
)

var (
	ErrDefault = ErrSystem
)
