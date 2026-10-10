<?php
echo preg_replace_callback('/x/',function($m){if(true)return $m[0];},'x');
