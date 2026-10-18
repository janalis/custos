<?php
function executableComments(PDO $db) {
    $mysql = $db->prepare('SELECT /*! ? */ ?');
    $mysql->execute([1, 2]);
    $maria = $db->prepare('SELECT /*M! ? */ ?');
    $maria->execute([1, 2]);
    $dash = $db->prepare('SELECT 1--? + ?');
    $dash->execute([1, 2]);
}
