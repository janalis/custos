<?php
// Before PHP 8.6 Override cannot target a constant at all: that is
// AttributeTargetMismatch's concern, not a missing target.
namespace Before0 { class C { #[\Override] const X = 1; } }
namespace Before1 { enum E { #[\Override] case A; } }
