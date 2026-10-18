<?php
function conditionalBinding(PDO $pdo, $flag) {
    $statement = $pdo->prepare('SELECT :id');
    if ($flag) { $statement->bindValue(':id', 7); }
    $statement->execute([]);
}
