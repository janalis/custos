<?php
function probedButNeverAdvanced(mysqli $db) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        do {
            if ($set = $db->store_result()) { $set->free(); }
        } while ($db->more_results());
        return <warning descr="Drain all multi-query results before another query.">$db->query('SELECT 14')</warning>;
    }
}
function otherConnectionDrained(mysqli $db, mysqli $replica) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        do {
        } while ($replica->next_result());
        return <warning descr="Drain all multi-query results before another query.">$db->query('SELECT 14')</warning>;
    }
}
function queryInsideDrainLoop(mysqli $db) {
    if ($db->multi_query('SELECT 7; SELECT 11')) {
        while ($db->next_result()) {
            <warning descr="Drain all multi-query results before another query.">$db->query('SELECT 14')</warning>;
        }
    }
}
