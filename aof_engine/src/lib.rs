use std::ffi::CStr;
use std::os::raw::c_char;

#[unsafe(no_mangle)]
pub extern "C" fn write_to_aof(command: *const c_char) {
    let c_str = unsafe {
        assert!(!command.is_null());
        CStr::from_ptr(command)
    };

    if let Ok(cmd_str) = c_str.to_str() {
        println!("AOF received command: {}", cmd_str);
    }
}
