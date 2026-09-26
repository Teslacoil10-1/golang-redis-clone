use std::ffi::CStr;
use std::fs::{self, File, OpenOptions};
use std::io::{BufWriter, Write};
use std::os::raw::c_char;
use std::sync::{Mutex, OnceLock};
use std::thread;
use std::time::{SystemTime, UNIX_EPOCH};

struct Manifest {
    base_file: Option<String>,
    active_incr: String,
}

struct AofEngine {
    dir: String,
    manifest: Manifest,
    buf: BufWriter<File>,
}

static ENGINE: OnceLock<Mutex<AofEngine>> = OnceLock::new();

fn init_engine() -> &'static Mutex<AofEngine> {
    ENGINE.get_or_init(|| {
        let dir = "/data/appendonlydir".to_string();
        fs::create_dir_all(&dir).unwrap();
        
        let initial_incr = format!("{}/appendonly.1.incr.aof", dir);
        let file = OpenOptions::new()
            .create(true)
            .append(true)
            .open(&initial_incr)
            .unwrap();
        
        Mutex::new(AofEngine {
            dir,
            manifest: Manifest {
                base_file: None,
                active_incr: initial_incr,
            },
            buf: BufWriter::with_capacity(8192, file),
        })
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn write_to_aof(command: *const c_char) {
    let c_str = unsafe {
        assert!(!command.is_null());
        CStr::from_ptr(command)
    };

    if let Ok(cmd_str) = c_str.to_str() {
        let engine_mu = init_engine();
        let mut engine = engine_mu.lock().unwrap();
        
        if let Err(e) = writeln!(engine.buf, "{}", cmd_str) {
            eprintln!("AOF_buf Write Error: {}", e);
        }
        

        let _ = engine.buf.flush();
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn trigger_bgrewriteaof() {
    let engine_mu = init_engine();
    let mut engine = engine_mu.lock().unwrap();
    
    let ts = SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_millis();
    let new_incr = format!("{}/appendonly.{}.incr.aof", engine.dir, ts);
    let new_base = format!("{}/appendonly.{}.base.aof", engine.dir, ts);
    
    let _ = engine.buf.flush();
    
    let file = OpenOptions::new()
        .create(true)
        .append(true)
        .open(&new_incr)
        .unwrap();
    
    engine.buf = BufWriter::with_capacity(8192, file);
    engine.manifest.active_incr = new_incr.clone();
    
    println!("swaped to new INCR file: {}", new_incr);

    thread::spawn(move || {
        println!("starting new subprocess for rewrite {}", new_base);
        
        let mut base_file = File::create(&new_base).unwrap();
        writeln!(base_file, "placeholder for db state snapshot").unwrap();
        base_file.sync_data().unwrap();
        
        println!("rewrite complete");
    });
}
