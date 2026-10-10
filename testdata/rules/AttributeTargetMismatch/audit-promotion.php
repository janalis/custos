<?php
#[Attribute(Attribute::TARGET_PROPERTY)] class Mark {}
class C {function __construct(#[Mark] public $x){}}
