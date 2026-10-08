<?php
// From PHP 8.0 Mailer() is an ordinary method, but DeprecatedConstructorStyle
// renames it to __construct: `: void` would make the combined fix fatal.
class Mailer
{
    /** @return void */
    public function Mailer($params = array())
    {
    }
}
