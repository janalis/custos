<?php

namespace Vendor\Auth;
class Token {}

namespace Acme\Tokens;
class Credential {}

namespace Acme\Tokens;

use Vendor\Auth\Token as Credential;
use Vendor\Auth\Token;

class Provider {}

$classes = [
    Provider::class,
    \Acme\Tokens\Credential::class,
    Token::class,
];
