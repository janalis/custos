<?php
function guard() {
    if (!current_user_can("x")) {
        wp_die("Sorry.");
    }
}
