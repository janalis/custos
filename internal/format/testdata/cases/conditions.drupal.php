<?php
// presets: per psr12 laravel drupal
if (
  $alpha &&
  $beta
) {
}
while (
  $a
  || $b
) {}
if (foo(
  $x
)) {}
$ok = $a /* why */ ||
  $b;
$fn =
  /** doc */
  static function () {};
$x = 0 % // c
  1;
if ($a) {
  foo();
  // describes the else
}
else {
  bar();
}
if ($a) {
  foo();
  // written at the brace
}
elseif ($b) {
  bar();
}
$a = $b
  ||
  $c;
$a = $b // why
  ||
  $c;
$x = 0 %
  // c
  1;
foo(
  $a,
  // the last argument is gone
);
switch ($a) {
  // first
  case 1:
    foo();

    // after the first case
  case 2:
    // after an empty case
  default:
    bar();
}
switch ($a):
    // first
    case 1:
endswitch;
enum Suit {

  case A;
  // a comment
  case B;

}
$ids = method_exists($failer, 'ids')
  ? $failer->ids($queue)
  : (new Collection($failer->all()))
    ->where('queue', $queue)
    ->toArray();
return $record instanceof MorphPivot
  ? $record->setMorphType($this->morphType)
    ->setMorphClass($this->morphClass)
  : $record;
$x = $foo
  ->bar();
(new Collection(static::$data))
  ->map(fn ($items) => (new Collection($items))
    ->map(function ($value) {
      return $value;
    }));
try {
  foo();
  // describes the catch
}
catch (E $e) {
}
return $this->pages()
  ->filter(fn ($page) => str_starts_with(
    $page,
    $search
  ))->keys()->first() ?? $this->pages()->map(fn ($page) => similar_text(
    $page,
    $search,
  ))
  ->filter(fn ($score) => $score >= 3)
  ->keys();
switch ($a) {
  case 1:
    foo();

    // falls through
  case 2:
    return 2;

  // describes the next case
  case 3:
    if ($x) {
      break;
    }

    // after a block
  default:
    bar();
}
