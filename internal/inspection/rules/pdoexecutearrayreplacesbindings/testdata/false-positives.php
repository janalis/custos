<?php
// The SQL has no placeholder: the empty array discards a stray binding.
function noMarker(PDO $pdo) { $s = $pdo->prepare('SELECT 1 AS v'); $s->bindValue(':id', 7); $s->execute([]); }
// The binding names a placeholder the SQL does not contain.
function absentMarker(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':other', 7); $s->execute([]); }
function mixedBindings(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); $s->bindValue(':other', 8); $s->execute([]); }
function positionOutOfRange(PDO $pdo) { $s = $pdo->prepare('SELECT ?'); $s->bindValue(2, 7); $s->execute([]); }
// The bound placeholder is not a known constant.
function unknownName(PDO $pdo, $name) { $s = $pdo->prepare('SELECT :id'); $s->bindValue($name, 7); $s->execute([]); }
// Nothing was bound, or values are supplied deliberately.
function noBinding(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->execute([]); }
function replaced(PDO $pdo) { $s = $pdo->prepare('SELECT :id'); $s->bindValue(':id', 7); $s->execute([':id' => 8]); }
