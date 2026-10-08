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
        if (<weak_warning descr="Compare with an empty string instead: '(string)$source === '''.">!strlen($source)</weak_warning>) {
            throw new RuntimeException('no source');
        }
        return <weak_warning descr="Compare with an empty string instead: '$native !== '''.">strlen($native) > 0</weak_warning>;
    }
}

/** @param string $label */
function hasLabel($label)
{
    return <weak_warning descr="Compare with an empty string instead: '(string)$label !== '''.">strlen($label) !== 0</weak_warning>;
}
