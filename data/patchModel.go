package data

type PatchModel struct {
	Increase  bool   `form:"Increase"`
	UpdatedBy string `form:"UpdatedBy"`
}
