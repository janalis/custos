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

function ternary_check(string $body)
{
    $data = json_decode($body, true);
    return (json_last_error() === JSON_ERROR_NONE) ? $data : $body;
}

function ternary_unrelated(string $body, bool $raw)
{
    $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
    return $raw ? $body : $data;
}

function closure_check(string $body)
{
    $data = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($body, true)</weak_warning>;
    return fn () => json_last_error() ? null : $data;
}

final class Parser
{
    private $sourceData;
    private $other;

    public function parse(string $response)
    {
        // The property is tested at its first read.
        $this->sourceData = json_decode($response, true);
        if (!$this->sourceData) {
            return null;
        }
        $this->other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($response, true)</weak_warning>;
        $this->other = [];
        $this->sourceData = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($response, true)</weak_warning>;
        return $this->sourceData['items'];
    }

    public function unread(string $response)
    {
        $this->other = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($response)</weak_warning>;
    }
}

final class Feed
{
    private $items;

    public function load(string $raw, string $name)
    {
        $$name = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw)</weak_warning>;
        $this->items = <weak_warning descr="Pass JSON_THROW_ON_ERROR in the flags of this call.">json_decode($raw, true)</weak_warning>;
        $late = function () {
            return $this->items ?: [];
        };
        return $late;
    }
}
