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
    $cast = (string) json_encode($data, JSON_THROW_ON_ERROR);
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
    $late = json_encode($v, JSON_THROW_ON_ERROR);
    echo $late;
    if (json_last_error()) {
        return null;
    }
    $other = json_encode($v, JSON_THROW_ON_ERROR);
    if ($other === '') {
        return null;
    }
    $last = json_encode($v, JSON_THROW_ON_ERROR);
    return [$rules, $late, $other, $last];
}

function unbraced(string $s)
{
    if ($s)
        $x = json_decode($s, false, 512, JSON_THROW_ON_ERROR);
    return $x;
}

function ternary_check(string $body)
{
    $data = json_decode($body, true);
    return (json_last_error() === JSON_ERROR_NONE) ? $data : $body;
}

function ternary_unrelated(string $body, bool $raw)
{
    $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
    return $raw ? $body : $data;
}

function closure_check(string $body)
{
    $data = json_decode($body, true, 512, JSON_THROW_ON_ERROR);
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
        $this->other = json_decode($response, true, 512, JSON_THROW_ON_ERROR);
        $this->other = [];
        $this->sourceData = json_decode($response, true, 512, JSON_THROW_ON_ERROR);
        return $this->sourceData['items'];
    }

    public function unread(string $response)
    {
        $this->other = json_decode($response, false, 512, JSON_THROW_ON_ERROR);
    }
}

final class Feed
{
    private $items;

    public function load(string $raw, string $name)
    {
        $$name = json_decode($raw, false, 512, JSON_THROW_ON_ERROR);
        $this->items = json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
        $late = function () {
            return $this->items ?: [];
        };
        return $late;
    }
}

function after_branch(string $out, int $mode)
{
    // The check after the `if` follows the decode ending its branch.
    if ($mode === 0) {
        $json = [];
    } elseif ($mode === 1) {
        $json = json_decode($out, true);
    } else {
        if ($mode === 2)
            $json = json_decode($out, false);
    }
    if (json_last_error() !== JSON_ERROR_NONE) {
        return ['error' => $out];
    }
    if ($mode) {
        $more = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
        echo $out;
    }
    if (json_last_error()) {
        return null;
    }
    while ($mode--) {
        $loop = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
    }
    if (json_last_error()) {
        return null;
    }
    if ($mode) {
        $last = json_decode($out, true, 512, JSON_THROW_ON_ERROR);
    }
    return [$json, $more, $loop, $last];
}
