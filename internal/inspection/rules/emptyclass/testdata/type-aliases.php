<?php

/**
 * The shape of one result row.
 *
 * @phpstan-type ResultRow array{path: string, line: int}
 */
final class ResultRowShape {}

/** @psalm-type Point = array{x: int, y: int} */
final class Geometry {}

/**
 * @phpstan-import-type ResultRow from ResultRowShape
 */
final class <weak_warning descr="This class declares no members; remove it or give it a purpose.">Importer</weak_warning> {}
