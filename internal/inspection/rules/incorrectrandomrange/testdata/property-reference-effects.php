<?php
class AliasedRange
{
    private $minimum = 9;
    private $maximum = 1;

    public function throughAlias()
    {
        $alias = &$this->minimum;
        changeRange($alias);
        return rand($this->minimum, $this->maximum);
    }

}

class MethodChangedRange
{
    private $minimum = 9;
    private $maximum = 1;

    public function throughMethod($writer)
    {
        $writer->fill($this->minimum);
        return rand($this->minimum, $this->maximum);
    }
}
