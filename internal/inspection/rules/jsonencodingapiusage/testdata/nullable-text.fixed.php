<?php

// Text that may be null decodes to null today; with JSON_THROW_ON_ERROR it
// would throw: reported without a fix.
class Account
{
    private $untyped;
    private ?string $typed = null;
    private string $text = '';
    private $defaulted = '[]';

    public function decode(?string $maybe, string $sure, \Other $other): array
    {
        return [
            json_decode($maybe, true),
            json_decode($this->untyped, true),
            json_decode($this->typed, true),
            json_decode($sure, true, 512, JSON_THROW_ON_ERROR),
            json_decode($this->text, true, 512, JSON_THROW_ON_ERROR),
            json_decode($this->defaulted, true, 512, JSON_THROW_ON_ERROR),
            json_decode($other->value, true, 512, JSON_THROW_ON_ERROR),
            json_decode($this->{'untyped'}, true, 512, JSON_THROW_ON_ERROR),
        ];
    }
}
