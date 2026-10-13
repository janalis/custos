<?php
final class Settings
{
    private array $values = [];

    /** @return string */
    public function <weak_warning descr="Declare ': string' as the return type.">label</weak_warning>()
    {
        // The value comes from an untyped source: only the doc says string,
        // so the declaration is suggested but not applied automatically.
        return $this->values['label'];
    }

    public function <weak_warning descr="Declare ': int' as the return type.">count</weak_warning>()
    {
        return \count($this->values);
    }
}
