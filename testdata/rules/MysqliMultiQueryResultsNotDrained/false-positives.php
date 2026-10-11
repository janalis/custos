<?php
function drainedByNextResult(mysqli $db) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        do {
            if ($set = $db->store_result()) { $set->free(); }
        } while ($db->next_result());
        return $db->query('SELECT 14');
    }
}
function drainedByBothProbes(mysqli $db) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        do {
            if ($set = $db->store_result()) { $set->free(); }
        } while ($db->more_results() && $db->next_result());
        return $db->query('SELECT 14');
    }
}
function drainedByWhile(mysqli $db) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        while ($db->next_result()) {
            if ($set = $db->store_result()) { $set->free(); }
        }
        return $db->query('SELECT 14');
    }
}
