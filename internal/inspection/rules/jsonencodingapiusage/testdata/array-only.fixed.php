<?php

// A failed decode gives null, which a parameter of PHP's own functions that
// only accepts an array rejects with a TypeError.
function keys(string $raw, array $base): array
{
    $a = array_keys(json_decode($raw, true));
    $b = array_merge($base, json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
    $c = count(json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
    $d = array_search('x', json_decode($raw, true));
    $e = strtoupper(json_decode($raw, false, 512, JSON_THROW_ON_ERROR));
    $f = mine(json_decode($raw, true));
    try {
        $g = array_keys(json_decode($raw, true, 512, JSON_THROW_ON_ERROR));
    } catch (\TypeError $e) {
        $g = [];
    }
    $item = json_decode($raw, true);
    $h = array_column($item['rows'], 'id');
    $i = array_keys(json_decode($raw, true)['rows']);
    $other = json_decode($raw, true, 512, JSON_THROW_ON_ERROR);
    $j = count($other['rows']);
    // A failed encode gives the empty text, which fails the decode it feeds.
    $k = array_keys(json_decode(json_encode($base), true));
    $l = json_decode((@json_encode($base)), true, 512, JSON_THROW_ON_ERROR);
    $m = json_decode(flags: JSON_THROW_ON_ERROR, json: json_encode($base));
    $n = json_decode(json_encode($base), true, 512, JSON_THROW_ON_ERROR);
    $o = array_keys(json_decode($raw, json_encode($base, JSON_THROW_ON_ERROR)));
    $p = strtoupper(json_encode($base, JSON_THROW_ON_ERROR));
    return [$a, $b, $c, $d, $e, $f, $g, $h, $i, $j, $k, $l, $m, $n, $o, $p];
}

function mine(array $a): array
{
    return $a;
}

function maybe(?array $a): array
{
    return $a ?? [];
}

// A parameter of the project declared `array` rejects null too.
class Settings
{
    public static function fromArray(array $values): self
    {
        return new self();
    }

    public function load(string $raw): array
    {
        return [
            self::fromArray(json_decode($raw, true)),
            Settings::fromArray(json_decode($raw, true)['nested']),
            mine(json_decode($raw, true)),
            maybe(json_decode($raw, true, 512, JSON_THROW_ON_ERROR)),
            static::fromArray(json_decode($raw, true, 512, JSON_THROW_ON_ERROR)),
            $this->fromArray(json_decode($raw, true, 512, JSON_THROW_ON_ERROR)),
        ];
    }
}
