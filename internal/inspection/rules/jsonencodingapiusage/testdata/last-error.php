<?php
function silenced($data)
{
    // The @ only hides the warning; the result is still tested.
    $json = @json_encode($data);
    if (false === $json) {
        return 'null';
    }
    if (!@json_decode($json)) {
        return null;
    }
    $cast = (string) <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($data)</weak_warning>;
    return [$json, $cast];
}

function last_error(string $s, $v)
{
    $rules = json_decode($s, true);
    if (JSON_ERROR_NONE !== json_last_error()) {
        throw new \InvalidArgumentException('bad');
    }
    @json_decode($s);
    if (\json_last_error_msg() !== 'No error') {
        return null;
    }
    $late = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v)</weak_warning>;
    echo $late;
    if (json_last_error()) {
        return null;
    }
    $other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v)</weak_warning>;
    if ($other === '') {
        return null;
    }
    $last = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_encode($v)</weak_warning>;
    return [$rules, $late, $other, $last];
}

function unbraced(string $s)
{
    if ($s)
        $x = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($s)</weak_warning>;
    return $x;
}
