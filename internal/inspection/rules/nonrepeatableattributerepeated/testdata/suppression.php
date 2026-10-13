<?php
// @custos-ignore NonRepeatableAttributeRepeated
#[Attribute] class Label {} #[Label, Label] class Parcel {}
