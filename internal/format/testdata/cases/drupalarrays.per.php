<?php
// presets: drupal per
$a = ['is_string', static function ($val) {
    return $val;
}];
$b = ['x' => 1, 'y' => [
    2,
]];
$c = ['one', 'line'];
$d = foo(1, function () {
    return 2;
});
$e = (new Foo())
    ->bar()
;
$f = ['http' => [
    // a comment inside
    'a' => 1,
]];
$g = [
    1, // after an item
    [
        2,
        3,
    ],
];
