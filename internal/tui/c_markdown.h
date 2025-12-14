#ifndef C_MARKDOWN_H
#define C_MARKDOWN_H

#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

// Markdown 解析器句柄
typedef struct MarkdownParser MarkdownParser;

// 创建新的 Markdown 解析器
MarkdownParser* markdown_parser_new(void);

// 释放 Markdown 解析器
void markdown_parser_free(MarkdownParser* parser);

// 解析 Markdown 文本并渲染为 ANSI 格式
// 输入：markdown_text - UTF-8 编码的 Markdown 文本
// 输出：返回动态分配的 ANSI 格式字符串，调用者需要使用 markdown_free_string() 释放
char* markdown_parse_to_ansi(MarkdownParser* parser, const char* markdown_text);

// 释放由 markdown_parse_to_ansi 返回的字符串
void markdown_free_string(char* s);

// 检查是否有错误发生
bool markdown_has_error(MarkdownParser* parser);

// 获取最后的错误信息
const char* markdown_get_error(MarkdownParser* parser);

// 设置渲染选项
void markdown_set_gfm_enabled(MarkdownParser* parser, bool enabled);
void markdown_set_table_enabled(MarkdownParser* parser, bool enabled);
void markdown_set_strikethrough_enabled(MarkdownParser* parser, bool enabled);
void markdown_set_tasklist_enabled(MarkdownParser* parser, bool enabled);

// 自定义颜色设置（ANSI 颜色代码）
void markdown_set_heading_color(MarkdownParser* parser, const char* color);
void markdown_set_code_color(MarkdownParser* parser, const char* color);
void markdown_set_link_color(MarkdownParser* parser, const char* color);
void markdown_set_text_color(MarkdownParser* parser, const char* color);

#ifdef __cplusplus
}
#endif

#endif // C_MARKDOWN_H