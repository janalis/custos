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
    <weak_warning descr="Use \Acme\Tokens\Provider::class instead of the class name string.">'Acme\Tokens\Provider'</weak_warning>,
    <weak_warning descr="Use \Acme\Tokens\Credential::class instead of the class name string.">'Acme\Tokens\Credential'</weak_warning>,
    <weak_warning descr="Use \Vendor\Auth\Token::class instead of the class name string.">'Vendor\Auth\Token'</weak_warning>,
];
