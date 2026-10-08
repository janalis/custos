<?php
class Widget
{
    /** @return string */
    public function option($name, $default = null)
    {
        return $default;
    }

    public function init(string $native)
    {
        $source = $this->option('source');
        if ((string)$source === '') {
            throw new RuntimeException('no source');
        }
        return $native !== '';
    }
}

/** @param string $label */
function hasLabel($label)
{
    return (string)$label !== '';
}
