<?php
#[Attribute(Attribute::TARGET_CLASS)] class Marker {} #[Marker] class Product {}

// Before PHP 8.5 a constant takes no attribute at all: the code is newer
// than the target, not a misplaced attribute.
#[Attribute] class Mark {}
#[Mark] const MARKED = 1;
