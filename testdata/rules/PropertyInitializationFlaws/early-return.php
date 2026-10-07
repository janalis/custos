<?php
class Key
{
    private array $data = [];
    private string $label = <weak_warning descr="Default is always replaced by the constructor; remove it.">'key'</weak_warning>;

    public function __construct(?array $data = null, string $label = 'k')
    {
        $this->label = $label;
        if (!$data) {
            return; // keeps the [] default
        }
        $this->data = $data;
    }
}

class Token
{
    private string $value = <weak_warning descr="Default is always replaced by the constructor; remove it.">''</weak_warning>;

    public function __construct(string $value)
    {
        $check = function () { return true; };
        $this->value = $value;
    }
}
