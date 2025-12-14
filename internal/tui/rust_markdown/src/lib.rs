use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use pulldown_cmark::{Parser, Options, Tag, TagEnd, Event, HeadingLevel};

// 解析器状态结构体
#[repr(C)]
pub struct MarkdownParser {
    error_message: *mut c_char,
    gfm_enabled: bool,
    table_enabled: bool,
    strikethrough_enabled: bool,
    tasklist_enabled: bool,
    heading_color: *mut c_char,
    code_color: *mut c_char,
    link_color: *mut c_char,
    text_color: *mut c_char,
}

// 创建新的 Markdown 解析器
#[no_mangle]
pub extern "C" fn markdown_parser_new() -> *mut MarkdownParser {
    let parser = Box::new(MarkdownParser {
        error_message: std::ptr::null_mut(),
        gfm_enabled: true,
        table_enabled: true,
        strikethrough_enabled: true,
        tasklist_enabled: true,
        heading_color: CString::new("86").unwrap().into_raw(),
        code_color: CString::new("252").unwrap().into_raw(),
        link_color: CString::new("39").unwrap().into_raw(),
        text_color: CString::new("255").unwrap().into_raw(),
    });
    Box::into_raw(parser)
}

// 释放 Markdown 解析器
#[no_mangle]
pub extern "C" fn markdown_parser_free(parser: *mut MarkdownParser) {
    if parser.is_null() {
        return;
    }
    
    unsafe {
        let parser = Box::from_raw(parser);
        
        // 释放字符串内存
        if !parser.error_message.is_null() {
            let _ = CString::from_raw(parser.error_message);
        }
        if !parser.heading_color.is_null() {
            let _ = CString::from_raw(parser.heading_color);
        }
        if !parser.code_color.is_null() {
            let _ = CString::from_raw(parser.code_color);
        }
        if !parser.link_color.is_null() {
            let _ = CString::from_raw(parser.link_color);
        }
        if !parser.text_color.is_null() {
            let _ = CString::from_raw(parser.text_color);
        }
    }
}

// 解析 Markdown 文本并渲染为 ANSI 格式
#[no_mangle]
pub extern "C" fn markdown_parse_to_ansi(
    parser: *mut MarkdownParser,
    markdown_text: *const c_char,
) -> *mut c_char {
    if parser.is_null() || markdown_text.is_null() {
        return std::ptr::null_mut();
    }
    
    let parser_ref = unsafe {
        &mut *parser
    };
    
    // 清除之前的错误
    if !parser_ref.error_message.is_null() {
        let _ = unsafe { CString::from_raw(parser_ref.error_message) };
        parser_ref.error_message = std::ptr::null_mut();
    }
    
    let markdown_cstr = unsafe { CStr::from_ptr(markdown_text) };
    let markdown = match markdown_cstr.to_str() {
        Ok(s) => s,
        Err(_) => {
            parser_ref.error_message = CString::new("Invalid UTF-8 input").unwrap().into_raw();
            return std::ptr::null_mut();
        }
    };
    
    // 配置解析选项
    let mut options = Options::empty();
    options.insert(Options::ENABLE_STRIKETHROUGH);
    options.insert(Options::ENABLE_TABLES);
    options.insert(Options::ENABLE_TASKLISTS);
    options.insert(Options::ENABLE_FOOTNOTES);
    
    // 创建解析器
    let parser_obj = Parser::new_ext(markdown, options);
    
    // 渲染为 ANSI
    let mut result = String::new();
    let mut list_stack = Vec::new();
    let mut list_index_stack = Vec::new();
    let mut in_code_block = false;
    
    for event in parser_obj {
        match event {
            Event::Start(tag) => {
                match tag {
                    Tag::Heading { level, .. } => {
                        // 确保标题前有换行
                        if !result.is_empty() && !result.ends_with('\n') {
                            result.push('\n');
                        }
                        result.push_str("\x1b[1;38;5;86m"); // 青色粗体
                        if level == HeadingLevel::H1 {
                            result.push_str("\x1b[4m"); // 一级标题加下划线
                        }
                    }
                    Tag::CodeBlock(_) => {
                        in_code_block = true;
                        // 确保代码块前有换行
                        if !result.is_empty() && !result.ends_with('\n') {
                            result.push('\n');
                        }
                        result.push_str("\x1b[48;5;236m\x1b[38;5;252m"); // 代码块样式
                    }
                    Tag::Emphasis => {
                        result.push_str("\x1b[3;38;5;204m"); // 斜体粉色
                    }
                    Tag::Strong => {
                        result.push_str("\x1b[1;38;5;203m"); // 粗体红色
                    }
                    Tag::Link { .. } => {
                        result.push_str("\x1b[4;38;5;39m"); // 蓝色下划线
                    }
                    Tag::List(ordered) => {
                        // 列表前确保有换行
                        if !result.is_empty() && !result.ends_with('\n') {
                            result.push('\n');
                        }
                        list_stack.push(ordered.is_some());
                        list_index_stack.push(1);
                    }
                    Tag::Item => {
                        // 列表项前添加缩进（根据嵌套层级）
                        let indent_level = list_stack.len().saturating_sub(1);
                        for _ in 0..indent_level {
                            result.push_str("  ");
                        }
                        if let Some(&is_ordered) = list_stack.last() {
                            if is_ordered {
                                let index = list_index_stack.last_mut().unwrap();
                                result.push_str(&format!("\x1b[38;5;78m{}. \x1b[0m", index));
                                *index += 1;
                            } else {
                                result.push_str("\x1b[38;5;78m• \x1b[0m");
                            }
                        }
                    }
                    Tag::BlockQuote(_) => {
                        if !result.is_empty() && !result.ends_with('\n') {
                            result.push('\n');
                        }
                        result.push_str("\x1b[3;38;5;245m│ "); // 灰色斜体引用，添加边框
                    }
                    Tag::Table(_) => {
                        result.push_str("\x1b[38;5;240m"); // 表格样式
                    }
                    Tag::TableHead => {
                        // 表头
                    }
                    Tag::TableRow => {
                        // 表行
                    }
                    Tag::TableCell => {
                        // 表格单元格
                        result.push_str(" | ");
                    }
                    Tag::Strikethrough => {
                        result.push_str("\x1b[9;38;5;240m"); // 删除线
                    }
                    Tag::Paragraph => {
                        // 段落开始 - 如果在列表中不添加额外换行
                        if list_stack.is_empty() && !result.is_empty() && !result.ends_with('\n') {
                            result.push('\n');
                        }
                    }
                    _ => {}
                }
            }
            Event::End(tag) => {
                match tag {
                    TagEnd::Heading(_) => {
                        result.push_str("\x1b[0m\n\n"); // 重置样式并添加换行
                    }
                    TagEnd::CodeBlock => {
                        in_code_block = false;
                        // 确保代码块内容后有换行
                        if !result.ends_with('\n') {
                            result.push('\n');
                        }
                        result.push_str("\x1b[0m\n"); // 重置样式并添加换行
                    }
                    TagEnd::Paragraph => {
                        // 段落结束 - 根据上下文添加换行
                        if list_stack.is_empty() {
                            result.push_str("\n\n"); // 普通段落后添加双换行
                        } else {
                            result.push('\n'); // 列表中的段落只添加单换行
                        }
                    }
                    TagEnd::Emphasis | TagEnd::Strong | TagEnd::Link => {
                        result.push_str("\x1b[0m"); // 重置样式
                    }
                    TagEnd::List(_) => {
                        list_stack.pop();
                        list_index_stack.pop();
                        // 顶层列表结束后添加额外换行
                        if list_stack.is_empty() {
                            result.push('\n');
                        }
                    }
                    TagEnd::Item => {
                        // 列表项结束后添加换行
                        if !result.ends_with('\n') {
                            result.push('\n');
                        }
                        result.push_str("\x1b[0m");
                    }
                    TagEnd::BlockQuote(_) => {
                        result.push_str("\x1b[0m\n\n");
                    }
                    TagEnd::Table => {
                        result.push_str("\x1b[0m\n\n");
                    }
                    TagEnd::TableRow => {
                        result.push_str(" |\n");
                    }
                    TagEnd::TableCell => {
                        // 单元格结束
                    }
                    TagEnd::Strikethrough => {
                        result.push_str("\x1b[0m");
                    }
                    _ => {}
                }
            }
            Event::Text(text) => {
                result.push_str(&text);
            }
            Event::Code(code) => {
                result.push_str(&format!("\x1b[48;5;237m\x1b[38;5;220m{}\x1b[0m", code));
            }
            Event::SoftBreak => {
                // 软换行 - 在代码块中保持为换行，普通文本中转为空格或换行
                if in_code_block {
                    result.push('\n');
                } else {
                    result.push(' ');
                }
            }
            Event::HardBreak => {
                result.push('\n');
            }
            Event::Rule => {
                result.push_str("\n────────────────────\n\n");
            }
            Event::TaskListMarker(checked) => {
                if checked {
                    result.push_str("\x1b[38;5;78m[✓]\x1b[0m ");
                } else {
                    result.push_str("\x1b[38;5;240m[ ]\x1b[0m ");
                }
            }
            Event::FootnoteReference(_) => {
                // 脚注引用
            }
            Event::Html(html) => {
                // HTML 内容
                result.push_str(&html);
            }
            Event::InlineHtml(html) => {
                // 行内 HTML
                result.push_str(&html);
            }
            Event::InlineMath(math) => {
                // 行内数学
                result.push_str(&format!("${}$", math));
            }
            Event::DisplayMath(math) => {
                // 块级数学
                result.push_str(&format!("$$\n{}\n$$", math));
            }
        }
    }
    
    // 清理多余的换行
    while result.ends_with("\n\n\n") {
        result.pop();
    }
    
    // 转换为 C 字符串
    match CString::new(result) {
        Ok(c_string) => c_string.into_raw(),
        Err(_) => {
            parser_ref.error_message = CString::new("Failed to create result string").unwrap().into_raw();
            std::ptr::null_mut()
        }
    }
}

// 检查是否有错误发生
#[no_mangle]
pub extern "C" fn markdown_has_error(parser: *mut MarkdownParser) -> bool {
    if parser.is_null() {
        return true;
    }
    
    unsafe {
        let parser_ref = &*parser;
        !parser_ref.error_message.is_null()
    }
}

// 获取最后的错误信息
#[no_mangle]
pub extern "C" fn markdown_get_error(parser: *mut MarkdownParser) -> *const c_char {
    if parser.is_null() {
        return std::ptr::null();
    }
    
    unsafe {
        let parser_ref = &*parser;
        parser_ref.error_message
    }
}

// 设置 GFM 选项
#[no_mangle]
pub extern "C" fn markdown_set_gfm_enabled(parser: *mut MarkdownParser, enabled: bool) {
    if !parser.is_null() {
        unsafe {
            (*parser).gfm_enabled = enabled;
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_table_enabled(parser: *mut MarkdownParser, enabled: bool) {
    if !parser.is_null() {
        unsafe {
            (*parser).table_enabled = enabled;
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_strikethrough_enabled(parser: *mut MarkdownParser, enabled: bool) {
    if !parser.is_null() {
        unsafe {
            (*parser).strikethrough_enabled = enabled;
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_tasklist_enabled(parser: *mut MarkdownParser, enabled: bool) {
    if !parser.is_null() {
        unsafe {
            (*parser).tasklist_enabled = enabled;
        }
    }
}

// 释放由 markdown_parse_to_ansi 返回的字符串
#[no_mangle]
pub extern "C" fn markdown_free_string(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            let _ = CString::from_raw(s);
        }
    }
}

// 设置颜色
fn set_color(parser: *mut MarkdownParser, color: *const c_char, field: &mut *mut c_char) {
    if parser.is_null() || color.is_null() {
        return;
    }
    
    unsafe {
        if !(*field).is_null() {
            let _ = CString::from_raw(*field);
        }
        
        let color_cstr = CStr::from_ptr(color);
        *field = CString::new(color_cstr.to_string_lossy().into_owned())
            .unwrap_or_else(|_| CString::new("255").unwrap())
            .into_raw();
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_heading_color(parser: *mut MarkdownParser, color: *const c_char) {
    if !parser.is_null() {
        unsafe {
            set_color(parser, color, &mut (*parser).heading_color);
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_code_color(parser: *mut MarkdownParser, color: *const c_char) {
    if !parser.is_null() {
        unsafe {
            set_color(parser, color, &mut (*parser).code_color);
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_link_color(parser: *mut MarkdownParser, color: *const c_char) {
    if !parser.is_null() {
        unsafe {
            set_color(parser, color, &mut (*parser).link_color);
        }
    }
}

#[no_mangle]
pub extern "C" fn markdown_set_text_color(parser: *mut MarkdownParser, color: *const c_char) {
    if !parser.is_null() {
        unsafe {
            set_color(parser, color, &mut (*parser).text_color);
        }
    }
}