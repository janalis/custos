<?php
#[Attribute(Attribute::TARGET_CLASS | Attribute::IS_REPEATABLE)] class Label {} #[Label, Label] class Parcel {}
