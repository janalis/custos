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
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($maybe, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($this->untyped, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($this->typed, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($sure, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($this->text, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($this->defaulted, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($other->value, true)</weak_warning>,
            <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($this->{'untyped'}, true)</weak_warning>,
        ];
    }
}
