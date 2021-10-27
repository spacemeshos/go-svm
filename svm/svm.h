#ifndef SVM_H
#define SVM_H

#include <stdarg.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>

/**
 * FFI representation for function result type.
 *
 * [`svm_result_t`] effectively has three variants:
 *
 * - Error variant.
 * - Receipt variant.
 * - No data, just okay state.
 *
 * Please note that [`svm_result_t`] implements [`std::ops::Try`], so you can
 * effectively use `?` everywhere and it will automatically return an
 * [`svm_result_t::new_error()`] if necessary.
 *
 * # Memory management
 *
 * All [`svm_result_t`] instances allocate memory using the system allocator,
 * so it's very easy to free contents from C and other languages.
 *
 * ```c, no_run
 * free(result->receipt);
 * free(result->error);
 * ```
 */
typedef struct svm_result_t {
  const uint8_t *receipt;
  const uint8_t *error;
  uint32_t buf_size;
} svm_result_t;

/**
 * Initializes the configuration options for all newly allocates SVM runtimes.
 */
struct svm_result_t svm_init(bool in_memory, const uint8_t *path, uint32_t path_len);

/**
 *
 * Start of the Public C-API
 *
 * * Each method is annotated with `#[no_mangle]`
 * * Each method has `unsafe extern "C"` before `fn`
 *
 * See `build.rs` for using `cbindgen` to generate `svm.h`
 *
 *
 * Creates a new SVM Runtime instance backed-by an in-memory KV.
 *
 * Returns it the created Runtime via the `runtime` parameter.
 *
 * # Examples
 *
 * ```rust
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 * ```
 *
 */
struct svm_result_t svm_runtime_create(void **runtime);

/**
 * Destroys the Runtime and its associated resources.
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * // Destroys the Runtime
 * unsafe { svm_runtime_destroy(runtime); }
 * ```
 *
 */
struct svm_result_t svm_runtime_destroy(void *runtime);

/**
 * Returns the number of currently allocated runtimes.
 */
void svm_runtimes_count(uint64_t *count);

/**
 * Validates syntactically a binary `Deploy Template` transaction.
 *
 * Should be called while the transaction is in the `mempool` of the Host.
 * In case the transaction isn't valid - the transaction should be discarded.
 *
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let message = b"message data...";
 * let _res = unsafe { svm_validate_deploy(runtime, message.as_ptr(), message.len() as u32) };
 * ```
 *
 */
struct svm_result_t svm_validate_deploy(void *runtime,
                                        const uint8_t *message,
                                        uint32_t message_size);

/**
 * Validates syntactically a binary `Spawn Account` transaction.
 *
 * Should be called while the transaction is in the `mempool` of the Host.
 * In case the transaction isn't valid - the transaction should be discarded.
 *
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let message = b"message data...";
 * let _res = unsafe { svm_validate_spawn(runtime, message.as_ptr(), message.len() as u32) };
 * ```
 *
 */
struct svm_result_t svm_validate_spawn(void *runtime,
                                       const uint8_t *message,
                                       uint32_t message_size);

/**
 * Validates syntactically a binary `Call Account` transaction.
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let message = b"message data...";
 * let _res = unsafe { svm_validate_call(runtime, message.as_ptr(), message.len() as u32) };
 * ```
 *
 */
struct svm_result_t svm_validate_call(void *runtime, const uint8_t *message, uint32_t message_size);

/**
 * Deploys a `Template`
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let envelope = b"envelope data...";
 * let message = b"message data...";
 * let context = b"context data...";
 *
 * let _res = unsafe {
 *   svm_deploy(
 *     runtime,
 *     envelope.as_ptr(),
 *     message.as_ptr(),
 *     message.len() as u32,
 *     context.as_ptr())
 * };
 * ```
 *
 */
struct svm_result_t svm_deploy(void *runtime,
                               const uint8_t *envelope,
                               const uint8_t *message,
                               uint32_t message_size,
                               const uint8_t *context);

/**
 * Spawns a new `Account`.
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let envelope = b"envelope data...";
 * let message = b"message data...";
 * let context = b"context data...";
 *
 * let _res = unsafe {
 *   svm_spawn(
 *     runtime,
 *     envelope.as_ptr(),
 *     message.as_ptr(),
 *     message.len() as u32,
 *     context.as_ptr())
 * };
 * ```
 *
 */
struct svm_result_t svm_spawn(void *runtime,
                              const uint8_t *envelope,
                              const uint8_t *message,
                              uint32_t message_size,
                              const uint8_t *context);

/**
 * Calls `verify` on an Account.
 * The inputs `envelope`, `message` and `context` should be the same ones
 * passed later to `svm_call`.(in case the `verify` succeeds).
 *
 * Returns the Receipt of the execution via the `receipt` parameter.
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let envelope = b"envelope data...";
 * let message = b"message data...";
 * let context = b"context data...";
 *
 * let _res = unsafe {
 *   svm_verify(
 *     runtime,
 *     envelope.as_ptr(),
 *     message.as_ptr(),
 *     message.len() as u32,
 *     context.as_ptr())
 * };
 * ```
 *
 */
struct svm_result_t svm_verify(void *runtime,
                               const uint8_t *envelope,
                               const uint8_t *message,
                               uint32_t message_size,
                               const uint8_t *context);

/**
 * `Call Account` transaction.
 * Returns the Receipt of the execution via the `receipt` parameter.
 *
 * # Examples
 *
 * ```rust, no_run
 * use svm_runtime_ffi::*;
 *
 * let mut runtime = std::ptr::null_mut();
 *
 * unsafe { svm_init(true, std::ptr::null(), 0); }
 *
 * let res = unsafe { svm_runtime_create(&mut runtime) };
 * assert!(res.is_ok());
 *
 * let envelope = b"envelope data...";
 * let message = b"message data...";
 * let context = b"context data...";
 *
 * let _res = unsafe {
 *   svm_call(
 *     runtime,
 *     envelope.as_ptr(),
 *     message.as_ptr(),
 *     message.len() as u32,
 *     context.as_ptr())
 * };
 * ```
 *
 */
struct svm_result_t svm_call(void *runtime,
                             const uint8_t *envelope,
                             const uint8_t *message,
                             uint32_t message_size,
                             const uint8_t *context);

struct svm_result_t svm_rewind(void *runtime, uint64_t layer_id);

struct svm_result_t svm_commit(void *runtime);

struct svm_result_t svm_get_account(void *runtime_ptr,
                                    const uint8_t *account_addr,
                                    uint64_t *balance,
                                    uint64_t *counter_upper_bits,
                                    uint64_t *counter_lower_bits,
                                    uint8_t *template_addr);

#endif /* SVM_H */
