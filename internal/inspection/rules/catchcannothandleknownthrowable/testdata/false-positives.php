<?php
// Evaluating the argument can raise something a catch clause handles.
try { throw new Error(match ($code) { 1 => 'one' }); } catch (UnhandledMatchError $e) {}
try { throw new Error($lazy->message); } catch (Exception $e) {}
try { throw new Error("failed: $lazy"); } catch (Exception $e) {}
try { throw new Error('failed: ' . $lazy); } catch (Exception $e) {}
try { throw new Error($rows[0]); } catch (Exception $e) {}
try { throw new Error(Codes::LABEL); } catch (Exception $e) {}
try { throw new Error(...$arguments); } catch (Exception $e) {}
try { throw new Error('failed', 0, clone $previous); } catch (Exception $e) {}

// The constructor may reject the arguments with a TypeError.
try { throw new Error($message); } catch (TypeError $e) {}
try { throw new Error('failed', 'code'); } catch (TypeError $e) {}
try { throw new Error('a', 1, null, 4); } catch (ArgumentCountError $e) {}
