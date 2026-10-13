<?php
// @custos-ignore AttributeTargetMismatch
#[Attribute(Attribute::TARGET_METHOD)] class Marker {} #[Marker] class Product {}
