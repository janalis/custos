<?php
final class Settings
{
    private array $values = [];

    /** @return string */
    public function label()
    {
        // The value comes from an untyped source: only the doc says string,
        // so the declaration is suggested but not applied automatically.
        return $this->values['label'];
    }

    public function count(): int
    {
        return \count($this->values);
    }
}
