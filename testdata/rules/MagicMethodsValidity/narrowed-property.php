<?php
class Edition
{
    public ?string $abbreviation = null;
    public ?string $title = null;

    public function __toString()
    {
        // narrowed by the truthiness check: a string
        if ($this->abbreviation) {
            return $this->abbreviation;
        }
        if (null === $this->title) {
            return '';
        }
        return $this->title;
    }
}

class Label
{
    public ?string $text = null;

    public function __toString()
    {
        <error descr="__toString must return string; got 'null'.">return $this->text;</error>
    }
}
