package tour

type ContentConsumeListener func(length int, resume bool)

var DevNullContentConsumeListener = func(length int, resume bool) {

}
