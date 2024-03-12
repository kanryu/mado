# What is IME?

In short, it is a program needed to input characters from various languages into a computer.

Yes, I don't know what you're talking about, right?
I think this is true from the perspective of people who speak most languages.
However, it is different for users of CJK (Chinese, Japanese, Korean) languages.

As you can easily see from the Unicode code table, CJK languages account for most of the tens of thousands of characters. Yes, we need thousands of characters to read and write our languages. It has been pointed out that there are approximately 3000 types of characters that are read and written on a daily basis.

Of course, it is not possible to accurately enter such a large number of characters using only about 100 keys on a keyboard. That's why we use IME applications on a daily basis.

This is not limited to Windows or Mac. All operating systems have IME support, including Linux desktop, Android, iOS, Chrome Book, etc. You just don't realize its existence.

## Experience using IME

Surprisingly, IME text input is achieved using almost the same GUI on so many OSes.

Simplifying the IME text environment, it consists of three elements.

1. Text (already entered)
2. Preedit (text being entered)
3. Candidates (token conversion list)

Of these, [1] is already input text, and is no different from the case where IME is not used.

When you turn on the IME and type a key, the text will be displayed in a special state. This is called [2] Preedit text, and it is displayed on the screen but is not actually posted to the Edit widget.

Preedit text is usually not finalized as is. This is because IME is used to enter characters that cannot normally be entered using the keyboard, so if you want to convert key input directly into characters, it is easier to turn IME off. Therefore, Preedit text needs to be converted to the desired token by the IME.

Preedit usually has multiple conversion candidates. This is because the tokens entered by the user through the IME are Hiragana in Japanese, Pinyin in Chinese, and Chamo (자모) in Korean, and these represent sounds and are different from specific characters. It's for a reason.

In English, you will often find words that have the same pronunciation but different spellings. In English, words can be written differently depending on their spelling, but in CJK languages input using IME, only the sounds are input using Preedit, and even the same sound can often have different letters or words. These are listed in the IME as conversion candidates [3]. 

Surprisingly, these three different elements appear as one unit to the user in a continuous coordinate system.

However, [2] and [3] are rendered by the IME, not by the application, and they just happen to be displayed at the same coordinates.

## Experiens for example

The simplest case is when you want to enter a sentence like the one below.

> [Fire alarm is spelled 火災報知器 in Japanese.]

First, enter this in the normal state with IME turned off.

> [Fire alarm is spelled ]

You enter the dedicated key on your keyboard and turn on the IME.

Normally, it is the "半角/全角" key in MS-Windows, and the "かな" key in Mac OS. These keys do not exist in the US layout, and alternative keys such as "Ctrl+Space" are input on such keyboards.

Note:
  that key code events will not post into the application until the user turns off the IME, including this key.
All of these keystrokes are used by the IME to select and confirm token.

>                       [かさいほうちき]
> [Fire alarm is spelled ]

You turned on the IME and entered the token.
You think you're typing "spelled" continuously on the screen, but technically it's displayed in a different IME Window than the Edit widget, and rendered superimposed on the same coordinates.

Note:
  To provide such a user experience, we need to accurately inform the IME of the current text cursor coordinates when the IME is turned on, at least up to the moment the user actually types preedit.

> [Fire alarm is spelled かさいほうちき]
>                       [火災]
>                       [香西]
>                       [川西]
>                       [kasai]

The tokens displayed from the second line onwards are Candidates, which are character strings that are conversion candidates for the token entered as preedit.

These are usually selected using Space, Tab or the up and down cursor keys.

You selected [火災] and pressed Enter. Then you will see a screen like this.

Note:
  At this time, the token [火災] is posted to the Window as a Char event.

>                          [ほうちき]
> [Fire alarm is spelled 火災]
>                          [報知器]
>                          [放置]
>                          [houchiki]

Candidates are still visible. The preedit for it is [ほうちき].
In other words, the token shown in Candidate may be part of preedit.
The remaining text will then be left as preedit.

You confirmed preedit again and turned IME off.

On Windows, press the "半角/全角" key again. On Mac, press another key called "英数".

> [Fire alarm is spelled 火災報知器]

No keycode events occur in the application window, including keycodes with IME turned OFF.
This is because it is used for IME work and it is rather harmful for the application to know the actual key code.

You should not trap keystrokes while typing in the IME unless you want to draw the exact same behavior as the IME Window yourself. It usually causes application malfunction.

Now, the user has turned off the IME. From now on, the user's keystrokes will trigger the Keycode event and Char event at the same time. In other words, it's the same as before.

