package utils

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

var (
	uni      *ut.UniversalTranslator
	trans    ut.Translator
	validate *validator.Validate
)

// 初始化验证器
func init() {
	// 创建翻译器
	zhTrans := zh.New()
	uni = ut.New(zhTrans, zhTrans)
	trans, _ = uni.GetTranslator("zh")

	// 获取验证器
	validate = binding.Validator.Engine().(*validator.Validate)
	// 注册翻译器
	_ = zh_translations.RegisterDefaultTranslations(validate, trans)

	// 注册自定义标签名称
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// 注册自定义验证器
	registerCustomValidators()
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators() {
	// 域名验证器
	_ = validate.RegisterValidation("domain", func(fl validator.FieldLevel) bool {
		domain := fl.Field().String()
		if domain == "" {
			return true
		}
		// 简单的域名验证，可以根据需要调整
		match, _ := regexp.MatchString(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`, domain)
		return match
	})

	// IP地址验证器
	_ = validate.RegisterValidation("ip", func(fl validator.FieldLevel) bool {
		ip := fl.Field().String()
		if ip == "" {
			return true
		}
		// 简单的IPv4验证，可以根据需要调整
		match, _ := regexp.MatchString(`^(\d{1,3}\.){3}\d{1,3}$`, ip)
		return match
	})

	// URL验证器
	_ = validate.RegisterValidation("url", func(fl validator.FieldLevel) bool {
		url := fl.Field().String()
		if url == "" {
			return true
		}
		// 简单的URL验证，可以根据需要调整
		match, _ := regexp.MatchString(`^(http|https)://[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*(\.[a-zA-Z]{2,})(:\d{1,5})?(/.*)?$`, url)
		return match
	})

	// 注册翻译
	_ = validate.RegisterTranslation("domain", trans, func(ut ut.Translator) error {
		return ut.Add("domain", "{0}不是有效的域名", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("domain", fe.Field())
		return t
	})

	_ = validate.RegisterTranslation("ip", trans, func(ut ut.Translator) error {
		return ut.Add("ip", "{0}不是有效的IP地址", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("ip", fe.Field())
		return t
	})

	_ = validate.RegisterTranslation("url", trans, func(ut ut.Translator) error {
		return ut.Add("url", "{0}不是有效的URL", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("url", fe.Field())
		return t
	})
}

// ValidateStruct 验证结构体
func ValidateStruct(obj interface{}) error {
	err := validate.Struct(obj)
	if err != nil {
		// 翻译错误信息
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return ValidationError("INVALID_VALIDATION", "无效的验证参数")
		}

		validationErrors := err.(validator.ValidationErrors)
		if len(validationErrors) > 0 {
			firstErr := validationErrors[0]
			fieldName := firstErr.Field()
			errorMsg := firstErr.Translate(trans)
			return ValidationError("VALIDATION_FAILED", errorMsg).WithDetails(map[string]string{
				"field": fieldName,
				"tag":   firstErr.Tag(),
			})
		}
	}
	return nil
}

// BindAndValidate 绑定请求参数并验证
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 根据Content-Type选择绑定方法
	contentType := c.GetHeader("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(obj); err != nil {
			return ValidationError("INVALID_JSON", "无效的JSON格式").WithError(err)
		}
	} else if strings.Contains(contentType, "multipart/form-data") {
		if err := c.ShouldBindWith(obj, binding.Form); err != nil {
			return ValidationError("INVALID_FORM", "无效的表单数据").WithError(err)
		}
	} else {
		if err := c.ShouldBind(obj); err != nil {
			return ValidationError("INVALID_REQUEST", "无效的请求数据").WithError(err)
		}
	}

	// 验证参数
	return ValidateStruct(obj)
}

// ValidateJSON 验证JSON请求
func ValidateJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return ValidationError("INVALID_JSON", "无效的JSON格式").WithError(err)
	}
	return ValidateStruct(obj)
}

// ValidateQuery 验证查询参数
func ValidateQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return ValidationError("INVALID_QUERY", "无效的查询参数").WithError(err)
	}
	return ValidateStruct(obj)
}

// ValidateForm 验证表单请求
func ValidateForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return ValidationError("INVALID_FORM", "无效的表单数据").WithError(err)
	}
	return ValidateStruct(obj)
}

// ValidateURI 验证URI参数
func ValidateURI(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		return ValidationError("INVALID_URI", "无效的URI参数").WithError(err)
	}
	return ValidateStruct(obj)
}

// ValidateHeader 验证请求头
func ValidateHeader(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindHeader(obj); err != nil {
		return ValidationError("INVALID_HEADER", "无效的请求头").WithError(err)
	}
	return ValidateStruct(obj)
}
